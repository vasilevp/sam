package config

type Config struct {
	Speed    byte
	Pitch    byte
	Mouth    byte
	Throat   byte
	Singmode bool
	Debug    bool
}

func DefaultConfig() *Config {
	return &Config{
		Speed:    72,
		Pitch:    64,
		Mouth:    128,
		Throat:   128,
		Singmode: false,
		Debug:    false,
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
	с.Singmode = true
}
