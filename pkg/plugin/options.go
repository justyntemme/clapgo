package plugin

import (
	"context"
	"fmt"
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/param"
	"github.com/justyntemme/clapgo/pkg/state"
)

// Option is a functional option for configuring plugins
type Option func(*PluginBase) error

// WithName sets the plugin name
func WithName(name string) Option {
	return func(p *PluginBase) error {
		p.Info.Name = name
		return nil
	}
}

// WithVendor sets the plugin vendor
func WithVendor(vendor string) Option {
	return func(p *PluginBase) error {
		p.Info.Vendor = vendor
		return nil
	}
}

// WithVersion sets the plugin version
func WithVersion(version string) Option {
	return func(p *PluginBase) error {
		p.Info.Version = version
		return nil
	}
}

// WithDescription sets the plugin description
func WithDescription(desc string) Option {
	return func(p *PluginBase) error {
		p.Info.Description = desc
		return nil
	}
}

// WithURL sets the plugin URL
func WithURL(url string) Option {
	return func(p *PluginBase) error {
		p.Info.URL = url
		return nil
	}
}

// WithManual sets the manual URL
func WithManual(manual string) Option {
	return func(p *PluginBase) error {
		p.Info.Manual = manual
		return nil
	}
}

// WithSupport sets the support URL
func WithSupport(support string) Option {
	return func(p *PluginBase) error {
		p.Info.Support = support
		return nil
	}
}

// WithFeatures adds plugin features
func WithFeatures(features ...string) Option {
	return func(p *PluginBase) error {
		p.Info.Features = append(p.Info.Features, features...)
		return nil
	}
}

// WithParameter adds a parameter to the plugin
func WithParameter(paramInfo param.Info) Option {
	return func(p *PluginBase) error {
		if p.ParamManager == nil {
			p.ParamManager = param.NewManager()
		}
		return p.ParamManager.Register(paramInfo)
	}
}

// WithParameters adds multiple parameters
func WithParameters(params ...param.Info) Option {
	return func(p *PluginBase) error {
		for _, paramInfo := range params {
			if err := WithParameter(paramInfo)(p); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithHost sets the host pointer
func WithHost(host unsafe.Pointer) Option {
	return func(p *PluginBase) error {
		p.InitWithHost(host)
		return nil
	}
}

// WithLogger sets a custom logger
func WithLogger(logger interface{}) Option {
	return func(p *PluginBase) error {
		// TODO: implement custom logger support
		return nil
	}
}

// WithStateVersion sets the state version for migration support
func WithStateVersion(version state.Version) Option {
	return func(p *PluginBase) error {
		if p.StateManager == nil {
			p.StateManager = state.NewManager(p.Info.ID, p.Info.Name, version)
		} else {
			// Update existing state manager with new version
			p.StateManager = state.NewManager(p.Info.ID, p.Info.Name, version)
		}
		return nil
	}
}

// WithSampleRate sets the initial sample rate with validation
func WithSampleRate(sampleRate float64) Option {
	return func(p *PluginBase) error {
		if sampleRate <= 0 {
			return ErrInvalidSampleRate
		}
		p.SampleRate = sampleRate
		return nil
	}
}

// WithValidation adds validation to the plugin options
func WithValidation(validate func(*PluginBase) error) Option {
	return func(p *PluginBase) error {
		return validate(p)
	}
}

// NewPluginWithOptions creates a new plugin base with options
func NewPluginWithOptions(id string, opts ...Option) (*PluginBase, error) {
	p := &PluginBase{
		Info: Info{
			ID:       id,
			Features: make([]string, 0),
		},
		SampleRate:   44100.0,
		IsActivated:  false,
		IsProcessing: false,
		ParamManager: param.NewManager(),
	}
	
	// Apply all options
	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, fmt.Errorf("failed to apply plugin option: %w", err)
		}
	}
	
	// Initialize state manager if not set
	if p.StateManager == nil {
		p.StateManager = state.NewManager(p.Info.ID, p.Info.Name, state.Version1)
	}
	
	return p, nil
}

// Builder provides a fluent API for plugin construction
type Builder struct {
	id      string
	options []Option
	errors  []error
}

// NewBuilder creates a new plugin builder
func NewBuilder(id string) *Builder {
	if id == "" {
		return &Builder{
			errors: []error{fmt.Errorf("plugin ID cannot be empty")},
		}
	}
	return &Builder{
		id:      id,
		options: make([]Option, 0),
		errors:  make([]error, 0),
	}
}

// WithName adds name to the builder
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errors = append(b.errors, fmt.Errorf("plugin name cannot be empty"))
		return b
	}
	b.options = append(b.options, WithName(name))
	return b
}

// WithVendor adds vendor to the builder
func (b *Builder) WithVendor(vendor string) *Builder {
	b.options = append(b.options, WithVendor(vendor))
	return b
}

// WithVersion adds version to the builder
func (b *Builder) WithVersion(version string) *Builder {
	b.options = append(b.options, WithVersion(version))
	return b
}

// WithDescription adds description to the builder
func (b *Builder) WithDescription(desc string) *Builder {
	b.options = append(b.options, WithDescription(desc))
	return b
}

// WithFeatures adds features to the builder
func (b *Builder) WithFeatures(features ...string) *Builder {
	b.options = append(b.options, WithFeatures(features...))
	return b
}

// WithParameter adds a parameter to the builder
func (b *Builder) WithParameter(paramInfo param.Info) *Builder {
	b.options = append(b.options, WithParameter(paramInfo))
	return b
}

// WithParameters adds multiple parameters to the builder
func (b *Builder) WithParameters(params ...param.Info) *Builder {
	b.options = append(b.options, WithParameters(params...))
	return b
}

// WithHost adds host configuration to the builder
func (b *Builder) WithHost(host unsafe.Pointer) *Builder {
	b.options = append(b.options, WithHost(host))
	return b
}

// WithStateVersion adds state version to the builder
func (b *Builder) WithStateVersion(version state.Version) *Builder {
	b.options = append(b.options, WithStateVersion(version))
	return b
}

// WithSampleRate adds sample rate to the builder
func (b *Builder) WithSampleRate(sampleRate float64) *Builder {
	if sampleRate <= 0 {
		b.errors = append(b.errors, ErrInvalidSampleRate)
		return b
	}
	b.options = append(b.options, WithSampleRate(sampleRate))
	return b
}

// WithValidation adds custom validation to the builder
func (b *Builder) WithValidation(validate func(*PluginBase) error) *Builder {
	b.options = append(b.options, WithValidation(validate))
	return b
}

// Build constructs the plugin with all configured options
func (b *Builder) Build() (*PluginBase, error) {
	// Check for builder errors first
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("builder has errors: %v", b.errors)
	}
	
	// Create plugin with options
	return NewPluginWithOptions(b.id, b.options...)
}

// BuildWithContext constructs the plugin with context support
func (b *Builder) BuildWithContext(ctx context.Context) (*PluginBase, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	// Build normally
	return b.Build()
}

// PresetConfig provides common plugin configurations

// AudioEffect creates a builder for audio effect plugins
func AudioEffect(id, name, vendor string) *Builder {
	return NewBuilder(id).
		WithName(name).
		WithVendor(vendor).
		WithVersion("1.0.0").
		WithFeatures("audio-effect", "stereo")
}

// Instrument creates a builder for instrument plugins  
func Instrument(id, name, vendor string) *Builder {
	return NewBuilder(id).
		WithName(name).
		WithVendor(vendor).
		WithVersion("1.0.0").
		WithFeatures("instrument", "stereo")
}

// Analyzer creates a builder for analyzer plugins
func Analyzer(id, name, vendor string) *Builder {
	return NewBuilder(id).
		WithName(name).
		WithVendor(vendor).
		WithVersion("1.0.0").
		WithFeatures("analyzer", "stereo")
}

// Example usage:
// plugin, err := AudioEffect("com.example.mygain", "My Gain", "Example Corp").
//     WithDescription("A simple gain plugin").
//     WithParameter(param.Volume(0, "Master")).
//     WithSampleRate(44100).
//     Build()
//
// Or with context:
// plugin, err := NewBuilder("com.example.myplugin").
//     WithName("My Plugin").
//     WithVendor("Example Corp").
//     WithValidation(func(p *PluginBase) error {
//         if len(p.Info.Features) == 0 {
//             return fmt.Errorf("plugin must have at least one feature")
//         }
//         return nil
//     }).
//     BuildWithContext(ctx)