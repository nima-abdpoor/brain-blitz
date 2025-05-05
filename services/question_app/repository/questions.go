package repository

import (
	"BrainBlitz.com/game/pkg/logger"
	"BrainBlitz.com/game/services/question_app/service"
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
)

type Config struct{}

type QuestionRepository struct {
	Config     Config
	Logger     logger.Logger
	PostgreSQL *sql.DB
}

func New(config Config, db *sql.DB, logger logger.Logger) QuestionRepository {
	return QuestionRepository{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}

func (repo QuestionRepository) GetProperQuestions(ctx context.Context, userIds []uint64, category []service.Category, limit int) ([]service.Question, error) {
	var questions []service.Question

	userIDArgs := make([]interface{}, len(userIds))
	for i, id := range userIds {
		userIDArgs[i] = id
	}

	seenQuery := `
		SELECT DISTINCT question_id 
		FROM user_question_history 
		WHERE user_id = ANY($1) 
		AND seen_at > NOW() - INTERVAL '30 days'
	`

	stmt, err := repo.PostgreSQL.PrepareContext(ctx, seenQuery)
	if err != nil {
		return questions, fmt.Errorf("failed to prepare seenQuery statement: %w", err)
	}
	rows, err := stmt.QueryContext(ctx, pq.Array(userIds))
	if err != nil {
		return questions, err
	}
	defer rows.Close()

	var seenQuestionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan question ID: %w", err)
		}
		seenQuestionIDs = append(seenQuestionIDs, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	categoryStrings := make([]string, len(category))
	for i, c := range category {
		categoryStrings[i] = string(c)
	}

	questionsQuery := `
		SELECT id, content, correct_answer, choices, category, difficulty 
		FROM questions 
		WHERE category = ANY($1)
		AND id != ALL($2)
		ORDER BY RANDOM()
		LIMIT $3
	`
	rows, err = repo.PostgreSQL.QueryContext(
		ctx,
		questionsQuery,
		pq.Array(categoryStrings),
		pq.Array(seenQuestionIDs),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var q service.Question
		err = rows.Scan(
			&q.Id,
			&q.Content,
			&q.CorrectAnswer,
			pq.Array(&q.Choices),
			&q.Category,
			&q.Difficulty,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func (repo QuestionRepository) GetRandomQuestions(ctx context.Context, category []service.Category, difficulty service.Difficulty, limit int) ([]service.Question, error) {
	var questions []service.Question

	questionsQuery := `
		SELECT id, content, correct_answer, choices, category, difficulty 
		FROM questions 
		WHERE category = ANY($1)
		AND difficulty = $2
		ORDER BY RANDOM()
		LIMIT $3
	`
	rows, err := repo.PostgreSQL.QueryContext(
		ctx,
		questionsQuery,
		pq.Array(category),
		difficulty,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var q service.Question
		err = rows.Scan(
			&q.Id,
			&q.Content,
			&q.CorrectAnswer,
			pq.Array(&q.Choices),
			&q.Category,
			&q.Difficulty,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}
		questions = append(questions, q)
	}

	return questions, nil
}
