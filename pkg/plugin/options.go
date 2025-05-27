package plugin

import (
	"unsafe"
	
	"github.com/justyntemme/clapgo/pkg/param"
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
			return nil, err
		}
	}
	
	return p, nil
}

// Example usage:
// plugin, err := NewPluginWithOptions("com.example.myplugin",
//     WithName("My Plugin"),
//     WithVendor("Example Corp"),
//     WithVersion("1.0.0"),
//     WithFeatures(FeatureAudioEffect, FeatureStereo),
//     WithParameter(param.Volume(0, "Master")),
//     WithParameter(param.Pan(1, "Pan")),
// )