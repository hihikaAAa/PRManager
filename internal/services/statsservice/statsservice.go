package statsservice

import(
	"context"

	"github.com/hihikaAAa/PRManager/internal/repository/postgres"
)

type StatsService struct{
	prRepo *postgres.PRRepository
}

func New(prRepo *postgres.PRRepository) *StatsService{
	return &StatsService{prRepo: prRepo}
}

type Stats struct {
    TotalPR int `json:"total_pr"`
    OpenPR int `json:"open_pr"`
    MergedPR int `json:"merged_pr"`
    Reviewers []ReviewerStat `json:"reviewers"`
}

type ReviewerStat struct {
    UserID string `json:"user_id"`
    Count int `json:"count"`
}

func (s *StatsService) GetStats(ctx context.Context)(Stats, error){
	raw, reviewers, err := s.prRepo.GetStats(ctx)
	if err != nil{
		return Stats{}, err
	}

	out := Stats{
		TotalPR: raw.TotalPR,
		OpenPR: raw.OpenPR,
		MergedPR: raw.MergedPR,
	}

	for _, r := range reviewers{
		out.Reviewers = append(out.Reviewers, ReviewerStat{
			UserID: r.UserID,
			Count: r.Count,
		})
	}

	return out,nil
}