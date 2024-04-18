package word

import (
	"strings"
	"unicode"
)

// Returns string that represents the current state of the game - how many letters are guessed
// Format: _ I N _ O _ S (WINDOWS), or for multiple words: T _ E   _ E _ T _ E R (THE   WEATHER)
func FormatSentenceGuessState(word string, guessedLetters map[string]bool) string {
	formattedWithUnderscores := ""
	for _, letter := range word {
		if string(letter) == " " {
			formattedWithUnderscores += "   "
		}
		val, guessed := guessedLetters[string(letter)]
		if guessed && val {
			formattedWithUnderscores += string(letter)
		} else {
			formattedWithUnderscores += "_ "
		}
	}
	return formattedWithUnderscores
}

func ProcessInput(word string, input string, guessedLetters map[string]bool, isWordGuessed *bool, message *string) {
	if IsInputLetter(input) {
		ProcessLetterInput(word, input, guessedLetters, isWordGuessed, message)
	} else {
		ProcessWordInput(word, input, isWordGuessed, message)
	}
}

func ProcessLetterInput(word string, letter string, guessedLetters map[string]bool, isWordGuessed *bool, message *string) {
	alreadyGuessed, inMap := guessedLetters[letter]

	if !inMap {
		RespondToWrongGuess("Invalid character entered! Make sure you only use english alphabet letters (a-z or A-Z)",
			isWordGuessed, message)
		return
	}

	if alreadyGuessed {
		RespondToWrongGuess("This letter has already been guessed. Wait for your turn and try again...",
			isWordGuessed, message)
		return
	}

	if IsLetterInAWord(word, letter) {
		guessedLetters[letter] = true

		if IsWordGuessedOnlyByLetters(word, guessedLetters) {
			*isWordGuessed = true
			*message = "You guessed the word/sentence correctly, congrats!"
			return
		}

		*isWordGuessed = false
		*message = "You correctly guessed the letter"
		return
	}

	RespondToWrongGuess("This letter is not in a word/sentence. Wait for your turn and try again", isWordGuessed, message)
}

func IsWordGuessedOnlyByLetters(word string, guessedLetters map[string]bool) bool {
	for _, letter := range word {
		guessed := guessedLetters[string(letter)]
		if !guessed {
			return false
		}
	}
	return true
}

// TODO: OVA FUNKCIJA BIRA REC IZ BAZE I PROVERAVA DA LI SE SVA SLOVA TE RECI NALAZE U MAPI SA KARAKTERI
func GenerateWord() {

}

func ProcessWordInput(word string, input string, isWordGuessed *bool, message *string) {
	if word == input {
		*isWordGuessed = true
		*message = "You guessed the word/sentence correctly, congrats!"
	} else {
		RespondToWrongGuess("Your guess does not match the word/sentence, wait for your move and try again",
			isWordGuessed, message)
	}
}

func RespondToWrongGuess(responseMessage string, isWordGuessed *bool, message *string) {
	*isWordGuessed = false
	*message = responseMessage
}

func IsInputLetter(input string) bool {
	return len(input) == 1 && unicode.IsLetter(rune(input[0]))
}

func IsLetterInAWord(word string, letter string) bool {
	return strings.Contains(word, letter)
}

func InitializeLetterMap() map[string]bool {
	letterMap := make(map[string]bool)
	for i := 'A'; i <= 'Z'; i++ {
		letterMap[string(i)] = false
	}
	return letterMap
}
