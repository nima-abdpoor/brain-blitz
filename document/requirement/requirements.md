# BrainBlitz

# UseCase

## User use-cases

### Register

user can register by email

### Login

user can log into application by email and password

## Game use-cases

Each game have starter user to initiate the game.  
a game can have questions of one or more categories.  
the starter can initiate: number of questions, time limit for answering each question, the difficulty level and the
category of the questions.  
each player should see their score.
game winner is the player who answered more correct questions.

# Entity

## User

- ID
- Email
- Avatar
- Password

## Game

- ID
- Category
- QuestionIDs
- PlayerIDs
- StartTime
- DurationTime

## Player

- ID
- UserID
- GameID
- Score
- Answers

## Question

- ID
- Question
- Answer List
- Correct Answers
- Difficulty
- Category

## Category

- ID
- Type
- Description