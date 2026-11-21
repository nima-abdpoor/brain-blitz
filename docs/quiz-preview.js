const questions = [
  {
    category: "Technology",
    question: "Which protocol carries live game updates in BrainBlitz?",
    options: ["SMTP", "WebSocket", "DNS", "FTP"],
    correct: 1,
  },
  {
    category: "Sports",
    question: "How many points is a basketball free throw worth?",
    options: ["One", "Two", "Three", "Four"],
    correct: 0,
  },
  {
    category: "Music",
    question: "How many strings does a standard violin have?",
    options: ["Five", "Six", "Three", "Four"],
    correct: 3,
  },
];

const title = document.querySelector("#preview-title");
const counter = document.querySelector("#quiz-counter");
const progressSegments = document.querySelectorAll(".preview-progress span");
const category = document.querySelector("#quiz-category");
const question = document.querySelector("#quiz-question");
const options = document.querySelector("#quiz-options");
const controls = document.querySelector("#quiz-controls");
const feedback = document.querySelector("#quiz-feedback");
const next = document.querySelector("#quiz-next");

let questionIndex = 0;
let score = 0;

function renderQuestion() {
  const item = questions[questionIndex];
  title.textContent = "Sample round";
  const position = String(questionIndex + 1).padStart(2, "0");
  const total = String(questions.length).padStart(2, "0");
  counter.textContent = `${position} / ${total}`;
  progressSegments.forEach((segment, index) => {
    segment.classList.toggle("is-current", index <= questionIndex);
  });
  category.textContent = item.category;
  question.textContent = item.question;
  question.parentElement.classList.remove("is-result");
  options.replaceChildren(...item.options.map((answer, index) => {
    const button = document.createElement("button");
    const letter = document.createElement("span");
    button.type = "button";
    button.className = "quiz-option";
    letter.className = "option-letter";
    letter.textContent = "ABCD"[index];
    letter.setAttribute("aria-hidden", "true");
    button.append(letter, answer);
    button.addEventListener("click", () => selectAnswer(button, index));
    return button;
  }));
  options.hidden = false;
  feedback.textContent = "Pick an answer to try it.";
  next.textContent = questionIndex === questions.length - 1 ? "Show score" : "Next question";
  next.disabled = true;
}

function selectAnswer(selected, selectedIndex) {
  if (selected.getAttribute("aria-disabled") === "true") return;

  const item = questions[questionIndex];
  const correct = selectedIndex === item.correct;
  if (correct) score += 1;

  [...options.children].forEach((button, index) => {
    button.setAttribute("aria-disabled", "true");
    button.tabIndex = button === selected ? 0 : -1;
    button.classList.toggle("is-correct", index === item.correct);
    button.classList.toggle("is-incorrect", button === selected && !correct);
  });
  feedback.textContent = correct ? "Correct!" : `Correct answer: ${item.options[item.correct]}.`;
  next.disabled = false;
}

function renderScore() {
  title.textContent = "Sample result";
  counter.textContent = "Complete";
  category.textContent = "Your score";
  question.textContent = `${score} / ${questions.length} correct`;
  question.parentElement.classList.add("is-result");
  options.hidden = true;
  feedback.textContent = "Sample round complete.";
  next.textContent = "Play again";
}

next.addEventListener("click", () => {
  questionIndex += 1;
  if (questionIndex > questions.length) {
    questionIndex = 0;
    score = 0;
  }
  if (questionIndex === questions.length) {
    renderScore();
  } else {
    renderQuestion();
  }
  question.focus({ preventScroll: true });
});

renderQuestion();
controls.hidden = false;
