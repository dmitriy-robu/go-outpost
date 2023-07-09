package provably_fair

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/model"
	"go-outpost/internal/lib/logger/sl"
	"go-outpost/internal/lib/random"
	"go-outpost/internal/repository"
	"golang.org/x/exp/slog"
	"math"
	"strconv"
	"time"
)

type ProvablyFairRandomizer struct {
	Algorithm  string
	ServerSeed string
	ClientSeed string
	Nonce      int
	Min        int
	Max        int
}

type ProvablyFair struct {
	ProvablyFairRandomizer *ProvablyFairRandomizer
	ProvablyFairRepository repository.ProvablyFairRepository
	log                    *slog.Logger
}

type ProvablyFairData struct {
	ClientSeed     string
	ServerSeed     string
	ServerHashSeed string
	Nonce          int
	Result         float64
	Min            int
	Max            int
}

func NewProvablyFair(
	ProvablyFairRepository repository.ProvablyFairRepository,
	log *slog.Logger,
) *ProvablyFair {
	return &ProvablyFair{
		ProvablyFairRandomizer: &ProvablyFairRandomizer{
			Algorithm:  "sha512",
			ServerSeed: random.NewRandomString(64),
			Nonce:      0,
		},
		ProvablyFairRepository: ProvablyFairRepository,
		log:                    log,
	}
}

func (f *ProvablyFair) GetRandomNumber(clientSeed string, maxProvability int) ProvablyFairData {
	serverSeed := random.NewRandomString(64)

	f.ProvablyFairRandomizer.Min = 0
	f.ProvablyFairRandomizer.Max = maxProvability
	f.ProvablyFairRandomizer.ClientSeed = clientSeed
	f.ProvablyFairRandomizer.ServerSeed = serverSeed

	result := f.getProvablyFairData()

	f.ProvablyFairRandomizer.Nonce++

	return result
}

func (f *ProvablyFair) getProvablyFairData() ProvablyFairData {
	h := hmac.New(sha512.New, []byte(f.ProvablyFairRandomizer.ServerSeed))
	h.Write([]byte(f.ProvablyFairRandomizer.ClientSeed + "-" + strconv.Itoa(f.ProvablyFairRandomizer.Nonce)))
	hash := hex.EncodeToString(h.Sum(nil))

	partOfHash := hash[:5]
	decimal, _ := strconv.ParseInt(partOfHash, 16, 64)

	result := math.Mod(float64(decimal), 10000) / 100

	return ProvablyFairData{
		ClientSeed:     f.ProvablyFairRandomizer.ClientSeed,
		ServerSeed:     f.ProvablyFairRandomizer.ServerSeed,
		ServerHashSeed: hash,
		Nonce:          f.ProvablyFairRandomizer.Nonce,
		Result:         result,
		Min:            f.ProvablyFairRandomizer.Min,
		Max:            f.ProvablyFairRandomizer.Max,
	}
}

func (f *ProvablyFair) StoreGameDraw(rouletteID int64, game config.Game) (int64, error) {
	const op = "ProvablyFair.StoreGameDraw"

	now := time.Now()

	gameDrawModel := &model.GameDraw{
		GameID:    rouletteID,
		Game:      game,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := f.ProvablyFairRepository.SaveGameDraw(*gameDrawModel)
	if err != nil {
		f.log.Error("failed to get number by color", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (f *ProvablyFair) StoreProvablyFair(data ProvablyFairData, drawID int64) error {
	const op = "ProvablyFair.StoreProvablyFair"

	now := time.Now()

	provablyFairModel := &model.ProvablyFair{
		GameDrawID:           drawID,
		ClientSeed:           data.ClientSeed,
		ServerSeed:           data.ServerSeed,
		ResultedHash:         data.ServerHashSeed,
		ResultedRandomNumber: data.Result,
		Min:                  0,
		Max:                  data.Max,
		Nonce:                data.Nonce,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err := f.ProvablyFairRepository.SaveProvablyFair(*provablyFairModel)
	if err != nil {
		f.log.Error("failed to get number by color", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
