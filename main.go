package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	spaces      []string    // the spaces where the curser can go
	cursor      int         // the digit representing the cursors position in the grid
	selected    map[int]int // the selected spaces
	columns     int
	rows        int
	turn        int // odd if player 1's turn, player 2 otherwise, starts at 1
	smallestDim int // smallest dimension, either rows value or columns value, depending on which is smaller
	win         int // 0 if no player has won, 1 for player 1 win, 2 for player 2
}

func initialModel() model {
	// initialize the columns and rows and dynamically create the array
	// max row and column size may be artificially limited by terminal size
	// recommended size, at least 3 and under 10, there is no upper limit but the grid dimensions must be at least 3
	columns := 3
	rows := 3
	spaces := make([]string, columns*rows)
	smallest := 3
	if rows < columns {
		smallest = rows
	} else {
		smallest = columns
	}

	// Initialize the spaces slice with empty strings
	for i := range spaces {
		spaces[i] = ""
	}
	return model{
		turn:        1,
		spaces:      spaces,
		selected:    make(map[int]int),
		columns:     columns,
		rows:        rows,
		smallestDim: smallest,
		win:         0,
	}
}

// we recursively check the neighbor
// direction is an int 0-7,
// 0 right, 1 right-down, 2 down, 3 down-left
// 4 left, 5 left-up, 6 up, 7 up-right
// index is the position we're checking in the selected map
// player is the player who's turn it is, 1 or 2
func (m model) checkNeigbor(direction int, index int, player int) int {
	// check that the current index is valid then call again to check next index
	var nextIndex int
	switch direction {
	case 0 | 1 | 7:
		nextIndex = index + 1
	case 3 | 4 | 5:
		nextIndex = index - 1
	}

	switch direction {
	case 1 | 2 | 3:
		nextIndex = index + m.columns
	case 5 | 6 | 7:
		nextIndex = index - m.columns
	}
	// check if the next index is valid
	if nextPlayer, ok := m.selected[nextIndex]; ok {
		// if it is check if its the correct player
		if nextPlayer == player {
			// next index is the correct player, so check the next neighbor
			return (m.checkNeigbor(direction, nextIndex, player) + 1)
		}
	}
	// base case
	return 1
}

// check all possible neighbors to see if there is a win
func (m *model) checkAllNeighbors(index int, player int) {
	diagonalLength := m.smallestDim
	var direction int
	win := false
	// first check if the winning set satisfies length condition for its direction
	switch direction {
	case 0:
		if m.checkNeigbor(direction, index, player) >= m.columns {
			win = true
		}
	case 1:
		if m.checkNeigbor(direction, index, player) >= diagonalLength {
			win = true
		}
	case 2:
		if m.checkNeigbor(direction, index, player) >= m.rows {
			win = true
		}
	case 3:
		if m.checkNeigbor(direction, index, player) >= diagonalLength {
			win = true
		}
	case 4:
		if m.checkNeigbor(direction, index, player) >= m.columns {
			win = true
		}
	case 5:
		if m.checkNeigbor(direction, index, player) >= diagonalLength {
			win = true
		}
	case 6:
		if m.checkNeigbor(direction, index, player) >= m.rows {
			win = true
		}
	case 7:
		if m.checkNeigbor(direction, index, player) >= diagonalLength {
			win = true
		}
	}
	if win && player == 1 {
		m.win = 1
	} else if win && player == 2 {
		m.win = 2
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" and "w" keys move the cursor up
		case "up", "k", "w":
			// block scrolling if user attempts to go above bounds
			if m.cursor > 0 && m.cursor-(m.columns-1) > 0 {
				m.cursor -= m.columns
			}

		// The "down" and "j" and "s" keys move the cursor down
		case "down", "j", "s":
			// block scrolling if user attempts to go below bounds
			if m.cursor+m.columns < len(m.spaces) {
				m.cursor += m.columns
			}
		// The "right" and "l" and "d" keys move the cursor right
		case "right", "l", "d":
			if m.cursor%m.columns != m.columns-1 {
				m.cursor++
			}

		// The "left" and "h" and "a" keys move the cursor left
		case "left", "h", "a":
			if m.cursor > 0 && (m.cursor%m.columns != 0) {
				m.cursor--
			}

		// first we need to check who attempted to make a turn and if the turn is valid
		// see rules at bottom for more details
		// we do this by checking if the new spot is selected in the old one

		// check rows and columns for matches
		// need to check if adjacent items are equal
		// we already know where there are selected spaces so we just need to see if there are any that align
		// for options := range m.selected {
		// 	// check all possible matches associated with options
		// }
		// then check all possible diagonals
		// length of diagonal wins are equal to the shortest length dimension

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			// if space doesnt have either player value in it
			if m.selected[m.cursor] != 1 && m.selected[m.cursor] != 2 {
				// figure out who's turn it is
				// if the turn number is divisible by 2, its player 2's turn
				if m.turn%2 == 0 {
					m.selected[m.cursor] = 2
					// check if win
					m.checkAllNeighbors(m.cursor, 2)
					// only increment turn only after valid turn is made
					m.turn += 1
				} else {
					// otherwise its player 1's turn
					m.selected[m.cursor] = 1
					m.checkAllNeighbors(m.cursor, 1)
					m.turn += 1
				}

			}
			// else {
			// 	// send message that user cannot override a space
			// 	fmt.Println("Error cannot move over another players space")
			// }

		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) View() string {

	// must have at least 3 rows and columns
	if m.columns < 3 || m.rows < 3 {
		os.Exit(1)
	}

	// The header
	s := "\n"
	// debug win display
	s += fmt.Sprint(m.win)
	s += "\n"
	// Iterate over our spaces
	for i, choice := range m.spaces {

		// Is the cursor pointing at this choice?
		cursor := "   " // no cursor
		if m.cursor == i {
			if m.turn%2 == 0 {
				cursor = "o >" // 2nd player cursor
			} else {
				cursor = "x >" // player 1 cursor
			}
		}

		// Is this choice selected?
		checked := " " // not selected
		// Make a selection
		if _, ok := m.selected[i]; ok {
			if m.selected[i] == 1 {
				checked = "x" // player 1 selected
			} else if m.selected[i] == 2 {
				checked = "o" // 2nd player selected
			}
		}

		// Render the spaces
		// we need to render 4 types of spaces
		// spaces who arent checked or have a cursor
		// spaces who have a cursor but not checked (the base case)
		// spaces who have a check but no cursor
		// spaces who both have a cursor and are checked

		// though since both cursor and checked have default space values
		// we dont need to add any special logic, since that is done above
		s += fmt.Sprintf("%s [%s] %s", cursor, checked, choice)

		if (i+1)%m.columns == 0 {
			s += "\n"
		}
	}

	if m.win == 1 {
		s += "\n Player 1 has won! \n"
	} else if m.win == 2 {
		s += "\n Player 2 has won! \n"
	} else if m.win == 0 && len(m.selected) == len(m.spaces) {
		s += "\nDraw! Try again!\n"
	}
	s += "\nPress q to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// Rules:
// player 1 goes first
// a player must select an EMPTY space on their turn
// the minimum grid size is 3
// a player must match the number of rows wide in a row or number of columns high in a row to win
// or a player can match any number of diagonal spaces in a row that is equal to or greater than the shortest dimension
// for example, if the grid is 20 wide but only 3 tall the players only need to mark 3 in a row diagonally

// bugs:
// where all selected spaces are x by default, or they're not the selection that the cursor is showing
// spaces can be changed

// missing features:
// playing multiple games
// cross game scoring
// checking if a player has won
// messages to tell you what you need to do, e.g. when the player attempts to make a move over a previously selected space, or a player has won
// ensure all algorithms work on larger game grids
