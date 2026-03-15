package gotmx

import (
	"fmt"
	"sync"
	"testing"
)

// Tests that registering a template with a name that is already taken returns an error
func TestRegisterReturnsErrorIfTemplateNameIsTaken(t *testing.T) {
	registry := NewTemplateRegistryDefault()
	err := registry.RegisterTemplate(NewStringLiteralTemplate(
		"my-template", "Hello World1", "dummy"))
	if err != nil {
		t.Error("Expected no error when registering first template")
	}
	err = registry.RegisterTemplate(NewStringLiteralTemplate(
		"my-template", "Hello World2", "dummy"))
	if err == nil {
		t.Error("Expected error when registering second template with same name")
	}
}

// Tests that if the template name is unique, we can access it with the namespace or without the namespace-
func TestGetTemplateWithOrWithoutNamespace(t *testing.T) {
	registry := NewTemplateRegistryDefault()
	err := registry.RegisterTemplate(NewStringLiteralTemplate(
		"my-template", "Hello World1", "dummy"))
	if err != nil {
		t.Error("Expected no error when registering first template")
	}

	template, err := registry.GetTemplate("my-template")
	if err != nil {
		t.Error("Expected template to exist, got error:", err)
	}
	if template == nil {
		t.Error("Expected template to not be nil")
	}
	_, err = registry.GetTemplate("dummy#my-template")
	if err != nil {
		t.Error("Expected template to exist, got error:", err)
	}
}

func TestNameCanBeReusedIfDifferentNamespace(t *testing.T) {
	registry := NewTemplateRegistryDefault()
	err := registry.RegisterTemplate(NewStringLiteralTemplate(
		"my-template", "Hello World1", "dummy"))
	if err != nil {
		t.Error("Expected no error when registering first template")
	}
	err = registry.RegisterTemplate(NewStringLiteralTemplate(
		"my-template", "Hello World2", "dummy2"))
	if err != nil {
		t.Error("Expected no error when registering second template")
	}

	// must cause an error if trying to access without namespace
	_, err = registry.GetTemplate("my-template")
	if err == nil {
		t.Error("Expected error when accessing ambiguous template without namespace")
	}

	_, err = registry.GetTemplate("dummy#my-template")
	if err != nil {
		t.Error("Expected template to exist, got error:", err)
	}
	_, err = registry.GetTemplate("dummy2#my-template")
	if err != nil {
		t.Error("Expected template to exist, got error:", err)
	}
}

// ================================================================================================
// ReplaceFrom tests
// ================================================================================================

func TestReplaceFromSwapsTemplates(t *testing.T) {
	live := NewTemplateRegistryDefault()
	_ = live.RegisterTemplate(NewStringLiteralTemplate("old-tmpl", "Old Content", "ns1"))

	fresh := NewTemplateRegistryDefault()
	_ = fresh.RegisterTemplate(NewStringLiteralTemplate("new-tmpl", "New Content", "ns2"))

	live.ReplaceFrom(fresh)

	// Old template should be gone
	_, err := live.GetTemplate("old-tmpl")
	if err == nil {
		t.Error("Expected old-tmpl to be gone after ReplaceFrom")
	}

	// New template should be available
	tmpl, err := live.GetTemplate("new-tmpl")
	if err != nil {
		t.Fatalf("Expected new-tmpl to exist after ReplaceFrom, got error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("Expected non-nil template")
	}
}

func TestReplaceFromSetsInitializedFlag(t *testing.T) {
	live := NewTemplateRegistryDefault()
	// initialized is false by default

	fresh := NewTemplateRegistryDefault()
	_ = fresh.RegisterTemplate(NewStringLiteralTemplate("tmpl", "Content", "ns"))

	live.ReplaceFrom(fresh)

	// After ReplaceFrom, initialized should be true, so GetTemplateExt
	// should not call Init() on the lazy loader
	if !live.initialized.Load() {
		t.Error("Expected initialized to be true after ReplaceFrom")
	}
}

func TestReplaceFromConcurrentReads(t *testing.T) {
	live := NewTemplateRegistryDefault()
	live.SetLazyTemplateLoader(noopLazyTemplateLoader{})
	_ = live.RegisterTemplate(NewStringLiteralTemplate("tmpl", "Original", "ns"))
	// Force initialization so reads don't trigger Init()
	live.initialized.Store(true)

	var wg sync.WaitGroup
	errs := make(chan error, 200)

	// Start concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Read should not panic even during swap
			_, _ = live.GetTemplate("tmpl")
		}()
	}

	// Perform swap concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			fresh := NewTemplateRegistryDefault()
			name := TemplateName(fmt.Sprintf("tmpl-%d", n))
			if err := fresh.RegisterTemplate(NewStringLiteralTemplate(name, "Content", "ns")); err != nil {
				errs <- err
			}
			live.ReplaceFrom(fresh)
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
