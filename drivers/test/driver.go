package test

import (
	"context"
	"path/filepath"

	"github.com/murlokswarm/app"
	"github.com/pkg/errors"
)

var (
	// ErrSimulated is an error that is set to simulated a return error
	// behavior.
	ErrSimulated = errors.New("simulated error")
)

// Driver is an app.Driver implementation for testing.
type Driver struct {
	app.BaseDriver

	SimulateErr   bool
	UseBaseDriver bool
	Ctx           context.Context
	Page          app.Page

	OnRun func()

	factory  app.Factory
	elements app.ElemDB
	dock     app.DockTile
	menubar  app.Menu
	uichan   chan func()
}

// Name satisfies the app.Driver interface.
func (d *Driver) Name() string {
	return "Test"
}

// Base satisfies the app.Driver interface.
func (d *Driver) Base() app.Driver {
	return d
}

// Run satisfies the app.Driver interface.
func (d *Driver) Run(f app.Factory) error {
	if d.SimulateErr {
		return ErrSimulated
	}

	d.factory = f

	elements := app.NewElemDB()
	elements = app.ConcurrentElemDB(elements)
	d.elements = elements

	d.dock = newDockTile(d)

	menubar, err := newMenu(d, "menu bar", app.MenuConfig{})
	if err != nil {
		return err
	}
	d.menubar = menubar

	d.uichan = make(chan func(), 256)

	if d.OnRun != nil {
		d.uichan <- d.OnRun
	}

	if d.Ctx == nil {
		return errors.New("driver.Ctx is nil")
	}

	end := false
	for {
		select {
		case <-d.Ctx.Done():
			if !end {
				close(d.uichan)
				end = true
			}

		case f := <-d.uichan:
			if f == nil {
				return nil
			}
			f()
		}
	}
}

// AppName satisfies the app.Driver interface.
func (d *Driver) AppName() string {
	return "Driver unit tests"
}

// Resources satisfies the app.Driver interface.
func (d *Driver) Resources(path ...string) string {
	resources := filepath.Join(path...)
	return filepath.Join("resources", resources)
}

// Storage satisfies the app.Driver interface.
func (d *Driver) Storage(path ...string) string {
	storage := filepath.Join(path...)
	return filepath.Join("storage", storage)
}

// NewWindow satisfies the app.Driver interface.
func (d *Driver) NewWindow(c app.WindowConfig) (app.Window, error) {
	if d.UseBaseDriver {
		return d.BaseDriver.NewWindow(c)
	}
	if d.SimulateErr {
		return nil, ErrSimulated
	}
	return newWindow(d, c)
}

// NewContextMenu satisfies the app.Driver interface.
func (d *Driver) NewContextMenu(c app.MenuConfig) (app.Menu, error) {
	if d.UseBaseDriver {
		return d.BaseDriver.NewContextMenu(c)
	}
	if d.SimulateErr {
		return nil, ErrSimulated
	}
	return newMenu(d, "context menu", c)
}

// NewPage satisfies the app.Driver interface.
func (d *Driver) NewPage(c app.PageConfig) error {
	if d.UseBaseDriver {
		return d.BaseDriver.NewPage(c)
	}
	if d.SimulateErr {
		return ErrSimulated
	}

	if d.Page != nil {
		d.Page.Close()
		d.Page = nil
	}

	page, err := newPage(d, c)
	if err != nil {
		return err
	}
	d.Page = page
	return nil
}

// NewTestPage satisfies the tests.PageTester interface.
func (d *Driver) NewTestPage(c app.PageConfig) (app.Page, error) {
	if err := d.NewPage(c); err != nil {
		return nil, err
	}
	return d.Page, nil
}

// Render satisfies the app.Driver interface.
func (d *Driver) Render(compo app.Component) error {
	if d.SimulateErr {
		return ErrSimulated
	}

	elem, err := d.elements.ElementByComponent(compo)
	if err != nil {
		return err
	}
	return elem.Render(compo)
}

// ElementByComponent satisfies the app.Driver interface.
func (d *Driver) ElementByComponent(c app.Component) (app.ElementWithComponent, error) {
	return d.elements.ElementByComponent(c)
}

// MenuBar satisfies the app.Driver interface.
func (d *Driver) MenuBar() (app.Menu, error) {
	if d.UseBaseDriver {
		return d.BaseDriver.MenuBar()
	}
	return d.menubar, nil
}

// Dock satisfies the app.Driver interface.
func (d *Driver) Dock() (app.DockTile, error) {
	if d.UseBaseDriver {
		return d.BaseDriver.Dock()
	}
	return d.dock, nil
}

// CallOnUIGoroutine satisfies the app.Driver interface.
func (d *Driver) CallOnUIGoroutine(f func()) {
	d.uichan <- f
}
