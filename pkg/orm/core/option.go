package core

type OpOpt func(o *Option)

type Option struct {
	WithFinalizer   bool
	WithSync        bool
	WithAllFields   bool
	WhenSpecChanged bool
}

func (o *Option) SetupOption(opts ...OpOpt) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithAllFields() OpOpt {
	return func(o *Option) {
		o.WithAllFields = true
	}
}

func WithFinalizer() OpOpt {
	return func(o *Option) {
		o.WithFinalizer = true
	}
}

func WithSync() OpOpt {
	return func(o *Option) {
		o.WithSync = true
	}
}

func WhenSpecChanged() OpOpt {
	return func(o *Option) {
		o.WhenSpecChanged = true
	}
}
