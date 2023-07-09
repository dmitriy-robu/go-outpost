package start

import (
	"fmt"
	"github.com/google/uuid"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/provably_fair"
	"go-outpost/internal/http-server/model"
	"go-outpost/internal/lib/logger/sl"
	"go-outpost/internal/repository"
	"golang.org/x/exp/slog"
)

type RouletteRoller struct {
	RouletteColors           map[config.Color]config.RouletteColorConfig
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
		RouletteColors:           config.RouletteWheelConfig.Colors,
		ProvablyFair:             ProvablyFair,
		RouletteWinnerRepository: RouletteWinnerRepository,
		log:                      log,
	}
}

type RouletteWinColorAndNumberData struct {
	Color  config.Color `json:"color"`
	Number int          `json:"number"`
}

type RouletteColorData struct {
	Color       config.Color `json:"color"`
	Probability float64      `json:"probability"`
}

func (r *RouletteRoller) Roll(roulette *model.Roulette) (*RouletteWinColorAndNumberData, error) {
	const op = "handlers.roulette.start.Roll"

	maxProbability := config.RouletteWheelConfig.MaxWinProbability

	clientSeed := uuid.New().String()

	provablyFairData := r.ProvablyFair.GetRandomNumber(clientSeed, maxProbability)
	stopAt := provablyFairData.Result

	colors := make(map[config.Color]RouletteColorData)
	for color, probability := range r.RouletteColors {
		colors[color] = RouletteColorData{
			Color:       color,
			Probability: probability.Probability,
		}
	}

	color := r.GetWinner(colors, stopAt)

	number, err := r.RouletteWinnerRepository.GetNumberByColor(color)
	if err != nil {
		r.log.Error("failed to get number by color", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = r.RouletteWinnerRepository.SaveWin(roulette, color, number); err != nil {
		r.log.Error("failed to save roulette winner", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	drawID, err := r.ProvablyFair.StoreGameDraw(roulette.ID, config.Roulette)
	if err != nil {
		r.log.Error("failed to store game draw", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = r.ProvablyFair.StoreProvablyFair(provablyFairData, drawID); err != nil {
		r.log.Error("failed to store provably fair", sl.Err(err))

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//r.handleWinners(roulette.ID, color)

	return &RouletteWinColorAndNumberData{
		Color:  color,
		Number: number,
	}, nil
}

func (r *RouletteRoller) GetWinner(colors map[config.Color]RouletteColorData, stopAt float64) config.Color {
	var (
		color              config.Color
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
