package start

import (
	"fmt"
	"github.com/google/uuid"
	config2 "go-outpost/internal/api/config"
	"go-outpost/internal/api/http-server/handlers/provably_fair"
	"go-outpost/internal/api/http-server/model"
	"go-outpost/internal/api/repository"
	"go-outpost/internal/lib/logger/sl"
	"golang.org/x/exp/slog"
)

type RouletteRoller struct {
	RouletteColors           map[config2.Color]config2.RouletteColorConfig
	ProvablyFair             *provably_fair.ProvablyFair
	RouletteWinnerRepository repository.RouletteWinnerRepository
	log                      *slog.Logger
}

func NewRouletteRoller(
	RouletteWinnerRepository repository.RouletteWinnerRepository,
	ProvablyFair *provably_fair.ProvablyFair,
	log *slog.Logger,
) *RouletteRoller {
	return &RouletteRoller{
		RouletteColors:           config2.RouletteWheelConfig.Colors,
		ProvablyFair:             ProvablyFair,
		RouletteWinnerRepository: RouletteWinnerRepository,
		log:                      log,
	}
}

type RouletteWinColorAndNumberData struct {
	Color  config2.Color `json:"color"`
	Number int           `json:"number"`
}

type RouletteColorData struct {
	Color       config2.Color `json:"color"`
	Probability float64       `json:"probability"`
}

func (r *RouletteRoller) Roll(roulette *model.Roulette) (*RouletteWinColorAndNumberData, error) {
	const op = "handlers.roulette.start.Roll"

	var (
		drawID           int64
		err              error
		number           int
		color            config2.Color
		colors           map[config2.Color]RouletteColorData
		stopAt           float64
		provablyFairData provably_fair.ProvablyFairData
		maxProbability   int
		clientSeed       string
		probability      config2.RouletteColorConfig
	)

	maxProbability = config2.RouletteWheelConfig.MaxWinProbability

	clientSeed = uuid.New().String()

	provablyFairData = r.ProvablyFair.GetRandomNumber(clientSeed, maxProbability)
	stopAt = provablyFairData.Result

	colors = make(map[config2.Color]RouletteColorData)
	for color, probability = range r.RouletteColors {
		colors[color] = RouletteColorData{
			Color:       color,
			Probability: probability.Probability,
		}
	}

	color = r.GetWinner(colors, stopAt)

	number, err = r.RouletteWinnerRepository.GetNumberByColor(color)
	if err != nil {
		r.log.Error("failed to get number by color", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = r.RouletteWinnerRepository.SaveWin(roulette, color, number); err != nil {
		r.log.Error("failed to save roulette winner", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	drawID, err = r.ProvablyFair.StoreGameDraw(roulette.ID, config2.Roulette)
	if err != nil {
		r.log.Error("failed to store game draw", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = r.ProvablyFair.StoreProvablyFair(provablyFairData, drawID); err != nil {
		r.log.Error("failed to store provably fair", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RouletteWinColorAndNumberData{
		Color:  color,
		Number: number,
	}, nil
}

func (r *RouletteRoller) GetWinner(colors map[config2.Color]RouletteColorData, stopAt float64) config2.Color {
	var (
		color              config2.Color
		currentProbability float64
	)

	for _, colorData := range colors {
		currentProbability += colorData.Probability

		if currentProbability >= stopAt {
			return colorData.Color
		}
	}

	return color
}
