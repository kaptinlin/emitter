package emitter

// Option configures an [Emitter] at construction.
//
// The Option type exists as a stable extension point: there are intentionally
// no options today. Future configuration knobs will be added here without
// changing the New signature.
type Option func(*config)

type config struct{}
