package config

type RouletteConfig struct {
	Colors            map[Color]RouletteColorConfig
	MaxWinProbability int
}

type RouletteColorConfig struct {
	Probability float64
	Multiplier  int
}

var RouletteWheelConfig = RouletteConfig{
	Colors: map[Color]RouletteColorConfig{
		Red: {
			Probability: 46.6,
			Multiplier:  2,
		},
		Black: {
			Probability: 46.6,
			Multiplier:  2,
		},
		Green: {
			Probability: 6.8,
			Multiplier:  14,
		},
	},
	MaxWinProbability: 100,
}
