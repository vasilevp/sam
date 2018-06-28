package config

type Config struct {
	Speed  byte `arg:"-s" help:"set speed value"`
	Pitch  byte `arg:"-p" help:"set pitch value"`
	Mouth  byte `arg:"-m" help:"set mouth value"`
	Throat byte `arg:"-t" help:"set throat value"`
	Sing   bool `arg:"-S" help:"enable singing mode (special treatment of pitch)"`
	Debug  bool `arg:"-D" help:"print additional debug messages"`
}

func DefaultConfig() *Config {
	return &Config{
		Speed:  72,
		Pitch:  64,
		Mouth:  128,
		Throat: 128,
		Sing:   false,
		Debug:  false,
	}
}

func (с *Config) SetSpeed(_speed byte) {
	с.Speed = _speed
}
func (с *Config) SetPitch(_pitch byte) {
	с.Pitch = _pitch
}
func (с *Config) SetMouth(_mouth byte) {
	с.Mouth = _mouth
}
func (с *Config) SetThroat(_throat byte) {
	с.Throat = _throat
}
func (с *Config) EnableSingmode() {
	с.Sing = true
}
