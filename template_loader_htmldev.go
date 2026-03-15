package gotmx

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// defaultDevDebounce is the default debounce duration for file change events.
const defaultDevDebounce = 1000 * time.Millisecond

type templateLoaderHTMLDev struct {
	watcher            *fsnotify.Watcher
	templateRegistry   TemplateRegistry
	timer              *time.Timer
	excludes           []string
	eagerExtensions    []string
	lazyExtensions     []string
	logger             Logger
	templateLoaderHtml *templateLoaderHTML
	rootPath           string
	debounce           time.Duration
	reloadCallback     func(err error)
}

func newTemplateLoaderHTMLDev(path string, excludes []string, eagerExtensions []string, lazyExtensions []string, templateRegistry TemplateRegistry) *templateLoaderHTMLDev {
	// Apply defaults if no extensions provided
	if len(eagerExtensions) == 0 {
		eagerExtensions = []string{".htm"}
	}
	if len(lazyExtensions) == 0 {
		lazyExtensions = []string{".html"}
	}

	htmlTemplateLoader := newTemplateLoaderHTML(os.DirFS(path), templateRegistry)
	htmlTemplateLoader.SetLogger(&noopLogger{})
	templateLoader := templateLoaderHTMLDev{
		templateRegistry:   templateRegistry,
		logger:             &noopLogger{},
		templateLoaderHtml: htmlTemplateLoader,
		rootPath:           path,
		excludes:           excludes,
		eagerExtensions:    eagerExtensions,
		lazyExtensions:     lazyExtensions,
		debounce:           defaultDevDebounce,
	}
	return &templateLoader
}

func (tl *templateLoaderHTMLDev) Load(namespace Namespace, _ TemplateName) error {
	return tl.templateLoaderHtml.Load(namespace, "")
}

func (tl *templateLoaderHTMLDev) Init() {
	if err := tl.Start(); err != nil {
		tl.logger.Error("Error watching templates", "error", err)
	}
}

func (tl *templateLoaderHTMLDev) SetLogger(logger Logger) {
	tl.logger = logger
	tl.templateLoaderHtml.SetLogger(logger)
}

// SetDebounce configures how long to wait after the last file change before reloading.
func (tl *templateLoaderHTMLDev) SetDebounce(d time.Duration) {
	tl.debounce = d
}

// SetReloadCallback sets a callback that is invoked after every reload attempt.
// If the reload succeeded, err is nil; otherwise it describes the failure.
func (tl *templateLoaderHTMLDev) SetReloadCallback(fn func(err error)) {
	tl.reloadCallback = fn
}

func (tl *templateLoaderHTMLDev) Start() error {

	// if the given filePath is a file, then we return an error
	fileInfo, err := os.Stat(tl.rootPath)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		return errors.New("given path must be a directory")
	}
	return tl.Reload()
}

// Reload performs an atomic template reload: it builds a fresh registry from disk,
// then swaps it into the live registry in a single operation. Concurrent readers
// never see an empty registry. If parsing fails, the previous templates remain intact.
func (tl *templateLoaderHTMLDev) Reload() error {
	tl.logger.Info("Reloading all templates")
	if tl.templateRegistry == nil {
		return errors.New("templateRegistry in templateLoaderHTMLDev is not set")
	}

	// Build a fresh registry in isolation
	freshRegistry := NewTemplateRegistryDefault()
	freshRegistry.SetLogger(tl.logger)
	freshLoader := newTemplateLoaderHTML(os.DirFS(tl.rootPath), freshRegistry)
	freshLoader.SetLogger(tl.logger)

	// Walk the directory tree, parsing all eager templates into the fresh registry
	err := tl.walkDirForReload(tl.rootPath, tl.excludes, freshLoader)
	if err != nil {
		tl.logger.Error("Error parsing templates during reload", "error", err)
		tl.notifyReloadCallback(err)
		// Keep previous templates intact — do not swap
		return err
	}

	// Atomic swap: replace the live registry contents in one operation.
	// If the live registry is a TemplateRegistryDefault, use the optimized ReplaceFrom.
	// Otherwise, fall back to clear + re-register (for custom registry implementations).
	if defaultReg, ok := tl.templateRegistry.(*TemplateRegistryDefault); ok {
		defaultReg.ReplaceFrom(freshRegistry)
	} else {
		tl.templateRegistry.ClearTemplates()
		// Re-parse directly into the live registry as fallback
		if walkErr := tl.walkDirForReload(tl.rootPath, tl.excludes, tl.templateLoaderHtml); walkErr != nil {
			tl.notifyReloadCallback(walkErr)
			return walkErr
		}
	}

	// Also update the loader used for lazy loading so it points to the live registry
	tl.templateLoaderHtml = newTemplateLoaderHTML(os.DirFS(tl.rootPath), tl.templateRegistry)
	tl.templateLoaderHtml.SetLogger(tl.logger)

	tl.notifyReloadCallback(nil)
	return tl.restartWatcher()
}

// notifyReloadCallback invokes the reload callback if one is set.
func (tl *templateLoaderHTMLDev) notifyReloadCallback(err error) {
	if tl.reloadCallback != nil {
		tl.reloadCallback(err)
	}
}

func (tl *templateLoaderHTMLDev) Stop() error {
	tl.logger.Info("Shutting down template watcher")
	if tl.watcher != nil {
		return tl.watcher.Close()
	}
	return nil
}

// restartWatcher closes the existing watcher and starts a new one.
func (tl *templateLoaderHTMLDev) restartWatcher() error {
	if tl.watcher != nil {
		if err := tl.watcher.Close(); err != nil {
			return err
		}
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	tl.watcher = watcher

	// Start listening for events on the watcher
	go func() {
		tl.handleFileWatcherEvents()
	}()

	return tl.walkDirForWatching(tl.rootPath, tl.excludes)
}

func (tl *templateLoaderHTMLDev) handleFileWatcherEvents() {
	for {
		select {
		case event, ok := <-tl.watcher.Events:
			if !ok {
				return
			}

			if tl.timer != nil {
				tl.timer.Stop()
			}

			tl.logger.Info("File event", "filePath", event.Name, "eventOperation", event.Op,
				"hasCreate", event.Has(fsnotify.Create),
				"hasWrite", event.Has(fsnotify.Write),
				"hasRemove", event.Has(fsnotify.Remove),
				"hasRename", event.Has(fsnotify.Rename),
				"hasChmod", event.Has(fsnotify.Chmod))
			tl.timer = time.AfterFunc(tl.debounce, func() {
				tl.reloadIfNecessary(event)
			})
		case err, ok := <-tl.watcher.Errors:
			if !ok {
				return
			}
			tl.logger.Error("Error while watching files", "error", err)
		}
	}
}

func (tl *templateLoaderHTMLDev) reloadIfNecessary(event fsnotify.Event) {
	// if one of our watched directories was removed, then we need to reload.
	// otherwise we assume a file has been removed and then we only reload, if it was a template file
	isWatchedDir := slices.Contains(tl.watcher.WatchList(), event.Name)
	ext := filepath.Ext(event.Name)
	isTemplateFile := tl.isTemplateExtension(ext)
	isCreateWriteRemoveOrRename := event.Has(fsnotify.Remove) || event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Rename)
	if isCreateWriteRemoveOrRename && (isWatchedDir || isTemplateFile) {
		if err := tl.Reload(); err != nil {
			tl.logger.Error("Error while reloading", "error", err)
		}
		return
	}
}

// isTemplateExtension returns true if the extension is configured as either eager or lazy
func (tl *templateLoaderHTMLDev) isTemplateExtension(ext string) bool {
	return slices.Contains(tl.eagerExtensions, ext) || slices.Contains(tl.lazyExtensions, ext)
}

// walkDirForReload walks the directory tree and parses eager templates into the given loader.
// It does NOT set up file watchers — that is handled separately by restartWatcher.
func (tl *templateLoaderHTMLDev) walkDirForReload(filePath string, excludes []string, loader *templateLoaderHTML) error {
	var walkFunc fs.WalkDirFunc = func(path string, info fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if tl.isExcluded(path, excludes) {
			return fs.SkipDir
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if slices.Contains(tl.eagerExtensions, ext) {
				return loader.loadFromFile(path)
			}
		}
		return nil
	}
	return filepath.WalkDir(filePath, walkFunc)
}

// walkDirForWatching walks the directory tree and adds directories to the file watcher.
func (tl *templateLoaderHTMLDev) walkDirForWatching(filePath string, excludes []string) error {
	var walkFunc fs.WalkDirFunc = func(path string, info fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if tl.isExcluded(path, excludes) {
			tl.logger.Debug("Excluding path", "path", path)
			return fs.SkipDir
		}
		if info.IsDir() {
			return tl.addWatchedDirectory(path)
		}
		return nil
	}
	return filepath.WalkDir(filePath, walkFunc)
}

func (tl *templateLoaderHTMLDev) isExcluded(path string, excludes []string) bool {

	// if the current path starts with any of the excluded paths, then we return true, otherwise false
	for _, exclude := range excludes {
		cleanFilePath := filepath.Clean(path)
		cleanExclude := filepath.Clean(exclude)
		if strings.HasPrefix(cleanFilePath, cleanExclude) {
			return true
		}
	}
	return false
}

func (tl *templateLoaderHTMLDev) addWatchedDirectory(path string) error {
	tl.logger.Info("Watching directory", "dir", path)
	return tl.watcher.Add(path)
}
