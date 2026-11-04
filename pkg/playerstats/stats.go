// Package playerstats defines the completed-game statistics shared by services.
package playerstats

type Statistics struct {
	GamesPlayed      int64   `json:"gamesPlayed" bson:"gamesPlayed"`
	Wins             int64   `json:"wins" bson:"wins"`
	Losses           int64   `json:"losses" bson:"losses"`
	Draws            int64   `json:"draws" bson:"draws"`
	TotalPoints      int64   `json:"totalPoints" bson:"totalPoints"`
	BestScore        int64   `json:"bestScore" bson:"bestScore"`
	AnswersSubmitted int64   `json:"answersSubmitted" bson:"answersSubmitted"`
	CorrectAnswers   int64   `json:"correctAnswers" bson:"correctAnswers"`
	WinRate          float64 `json:"winRate"`
	AveragePoints    float64 `json:"averagePoints"`
	Accuracy         float64 `json:"accuracy"`
	Rank             int64   `json:"rank" bson:"rank"`
	RankedPlayers    int64   `json:"rankedPlayers" bson:"rankedPlayers"`
}

func (s *Statistics) CalculateRates() {
	if s.GamesPlayed > 0 {
		s.WinRate = 100 * float64(s.Wins) / float64(s.GamesPlayed)
		s.AveragePoints = float64(s.TotalPoints) / float64(s.GamesPlayed)
	}
	if s.AnswersSubmitted > 0 {
		s.Accuracy = 100 * float64(s.CorrectAnswers) / float64(s.AnswersSubmitted)
	}
}
