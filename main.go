package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	menuScreen screen = iota
	betScreen
	gameScreen
)

type suit int

const (
	clubs suit = iota
	diamonds
	hearts
	spades
)

type card struct {
	rank string
	suit suit
}

func (c card) points() int {
	switch c.rank {
	case "A":
		return 11
	case "K", "Q", "J":
		return 10
	default:
		var n int
		_, _ = fmt.Sscanf(c.rank, "%d", &n)
		return n
	}
}

func (c card) symbol() string {
	return []string{"♣", "♦", "♥", "♠"}[c.suit]
}

func (c card) color() lipgloss.Color {
	if c.suit == diamonds || c.suit == hearts {
		return lipgloss.Color("203")
	}
	return lipgloss.Color("236")
}

func (c card) String() string { return c.rank + c.symbol() }

type model struct {
	screen       screen
	menu         int
	deck         []card
	player       []card
	dealer       []card
	finished     bool
	message      string
	messageColor lipgloss.Color
	blackjack    bool
	bankroll     int
	bet          int
	betCursor    int
	width        int
	height       int
}

func newModel() model {
	return model{screen: menuScreen, menu: 0, bankroll: 100, bet: 10, width: 80, height: 24}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" && m.screen == menuScreen {
			return m, tea.Quit
		}
		switch m.screen {
		case menuScreen:
			switch key {
			case "j", "down":
				m.menu = (m.menu + 1) % 2
			case "k", "up":
				m.menu = (m.menu + 1) % 2
			case "enter", " ", "h":
				if m.menu == 0 {
					// Starting from the main menu begins a fresh session.
					m.bankroll = 100
					m.bet = 10
					m.screen = betScreen
					m.betCursor = m.bet/5 - 1
				} else {
					return m, tea.Quit
				}
			}
		case betScreen:
			if m.bankroll < 5 {
				switch key {
				case "enter", "esc":
					m.screen = menuScreen
				case "h":
					m.bankroll = 100
					m.bet = 10
					m.betCursor = 1
					m.screen = betScreen
				}
				return m, nil
			}
			maxBet := m.bankroll / 5
			switch key {
			case "k", "up":
				if m.betCursor < maxBet-1 {
					m.betCursor++
				}
			case "j", "down":
				if m.betCursor > 0 {
					m.betCursor--
				}
			case "enter", " ", "h":
				m.bet = min((m.betCursor+1)*5, m.bankroll)
				m.startGame()
			case "esc", "m":
				m.screen = menuScreen
			}
		case gameScreen:
			if m.finished {
				if key == "enter" || key == "h" {
					if m.bankroll < 5 {
						m.screen = betScreen
						return m, nil
					}
					m.screen = betScreen
					m.betCursor = min(m.bet/5-1, m.bankroll/5-1)
					if m.betCursor < 0 {
						m.betCursor = 0
					}
				} else if key == "esc" {
					m.screen = menuScreen
				}
				return m, nil
			}
			switch key {
			case "h":
				m.hit()
			case "j", "k", "l", "left", "down", "up", "right":
				// These directional keys are reserved for navigating future panels.
			case "s", " ":
				m.stand()
			case "esc", "m":
				m.screen = menuScreen
			}
		}
	}
	return m, nil
}

func (m *model) startGame() {
	if m.bet < 5 || m.bet > m.bankroll {
		m.bet = min(10, m.bankroll)
		if m.bet < 5 {
			m.bet = 5
		}
	}
	m.deck = makeDeck()
	rand.Shuffle(len(m.deck), func(i, j int) { m.deck[i], m.deck[j] = m.deck[j], m.deck[i] })
	m.player, m.dealer = nil, nil
	m.finished = false
	m.message = ""
	m.blackjack = false
	m.messageColor = gold
	m.bankroll -= m.bet
	m.screen = gameScreen
	m.player = append(m.player, m.draw(), m.draw())
	m.dealer = append(m.dealer, m.draw(), m.draw())
	if score(m.player) == 21 {
		m.blackjack = true
		m.finished = true
		m.revealDealer()
		if score(m.dealer) == 21 {
			m.message = "PUSH — both have blackjack."
			m.bankroll += m.bet
			m.messageColor = gold
		} else {
			m.message = "BLACKJACK! You win 3:2 (rounded down) "
			winnings := m.bet * 3 / 2
			winnings = winnings / 5 * 5
			m.bankroll += m.bet + winnings
			m.messageColor = lipgloss.Color("39")
		}
	}
}

func makeDeck() []card {
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
	deck := make([]card, 0, 52)
	for s := clubs; s <= spades; s++ {
		for _, rank := range ranks {
			deck = append(deck, card{rank: rank, suit: s})
		}
	}
	return deck
}

func (m *model) draw() card {
	c := m.deck[len(m.deck)-1]
	m.deck = m.deck[:len(m.deck)-1]
	return c
}

func (m *model) hit() {
	m.player = append(m.player, m.draw())
	if score(m.player) > 21 {
		m.finished = true
		m.message = "BUST — the house takes this hand."
		m.messageColor = lipgloss.Color("196")
	}
}

func (m *model) stand() {
	m.revealDealer()
	m.finished = true
	ps, ds := score(m.player), score(m.dealer)
	switch {
	case ps > 21:
		m.message = "BUST — the house takes this hand."
		m.messageColor = lipgloss.Color("196")
	case ds > 21:
		m.message = "DEALER BUST — you win!"
		m.messageColor = lipgloss.Color("42")
		m.bankroll += m.bet * 2
	case ps > ds:
		m.message = "YOU WIN — a better hand."
		m.messageColor = lipgloss.Color("42")
		m.bankroll += m.bet * 2
	case ps < ds:
		m.message = "DEALER WINS — better luck next hand."
		m.messageColor = lipgloss.Color("196")
	default:
		m.message = "PUSH — it’s a tie."
		m.messageColor = gold
		m.bankroll += m.bet
	}
}

func (m *model) revealDealer() {
	for score(m.dealer) < 17 {
		m.dealer = append(m.dealer, m.draw())
	}
}

func score(hand []card) int {
	total, aces := 0, 0
	for _, c := range hand {
		total += c.points()
		if c.rank == "A" {
			aces++
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

func (m model) View() string {
	switch m.screen {
	case menuScreen:
		return m.menuView()
	case betScreen:
		return m.betView()
	case gameScreen:
		return m.gameView()
	default:
		return ""
	}
}

var (
	green    = lipgloss.Color("42")
	gold     = lipgloss.Color("220")
	muted    = lipgloss.Color("245")
	cream    = lipgloss.Color("230")
	brand    = lipgloss.NewStyle().Bold(true).Foreground(gold)
	label    = lipgloss.NewStyle().Foreground(muted).Bold(true)
	cardFace = lipgloss.NewStyle().Foreground(cream).Background(lipgloss.Color("252")).Bold(true)
	cardBack = lipgloss.NewStyle().Foreground(lipgloss.Color("156")).Background(lipgloss.Color("22")).Bold(true)
)

func (m model) menuView() string {
	title := brand.Render("♠  BLACKJACK  ♣")
	subtitle := lipgloss.NewStyle().Foreground(muted).Render("A quiet table. One deck. Your call.")
	items := []string{"Start a new game", "Exit"}
	var menu strings.Builder
	for i, item := range items {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(cream)
		if m.menu == i {
			prefix = lipgloss.NewStyle().Foreground(gold).Bold(true).Render("› ")
			style = style.Bold(true).Foreground(gold)
		}
		menu.WriteString(prefix + style.Render(item) + "\n")
	}
	keys := lipgloss.JoinHorizontal(lipgloss.Left,
		helpLine("j / k", "move"), "     ", helpLine("enter / h", "select"), "     ", helpLine("q", "quit"))
	content := lipgloss.JoinVertical(lipgloss.Center, title, subtitle, "", menu.String(), keys)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(2, 5).Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) betView() string {
	if m.bankroll < 5 {
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("GAME OVER")
		message := lipgloss.NewStyle().Foreground(muted).Render("You are out of funds...")
		keys := helpLine("enter / esc", "menu", "h", "play again")
		content := lipgloss.JoinVertical(lipgloss.Center, title, message, "", keys)
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(2, 4).Align(lipgloss.Center).Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	maxBet := m.bankroll / 5
	if maxBet < 1 {
		maxBet = 1
	}
	title := brand.Render("♠  PLACE YOUR BET  ♣")
	bank := lipgloss.NewStyle().Foreground(muted).Render(fmt.Sprintf("Bankroll: $%d", m.bankroll))
	amount := min((m.betCursor+1)*5, m.bankroll)
	betLine := lipgloss.NewStyle().Bold(true).Foreground(green).Render(fmt.Sprintf("Bet: $%d", amount))
	keys := lipgloss.JoinVertical(lipgloss.Center,
		helpLine("j", "decrease bet by $5", "k", "increase bet by $5"),
		helpLine("enter / h", "deal", "esc", "menu"),
	)
	content := lipgloss.JoinVertical(lipgloss.Center, title, bank, "", betLine, "", keys)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(1, 4).Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m model) gameView() string {
	title := brand.Render("♠  BLACKJACK  ♣")
	dealerLabel := label.Render("DEALER")
	playerLabel := label.Render("YOUR HAND")
	dealerCards := renderHand(m.dealer, m.finished)
	playerCards := renderHand(m.player, true)
	dscore := "?"
	if m.finished {
		dscore = fmt.Sprint(score(m.dealer))
	}
	details := lipgloss.NewStyle().Foreground(muted).Render("Dealer total: " + dscore + "\nYour total: " + fmt.Sprint(score(m.player)))
	controls := lipgloss.JoinHorizontal(lipgloss.Left, helpLine("h", "hit"), "     ", helpLine("s", "stand"), "     ", helpLine("esc", "menu"))
	if m.finished {
		controls = lipgloss.JoinHorizontal(lipgloss.Left,
			helpLine("enter / h", "deal again"), "     ", helpLine("esc", "menu"))
	}
	controlLine := controls
	bankrollLine := lipgloss.NewStyle().Foreground(muted).Render(fmt.Sprintf("Bankroll: $%d     Current bet: $%d", m.bankroll, m.bet))
	parts := []string{title, bankrollLine, "", dealerLabel, dealerCards, details, "", playerLabel, playerCards, ""}
	if m.finished {
		parts = append(parts, lipgloss.NewStyle().Bold(true).Foreground(m.messageColor).Render(m.message), "")
	}
	parts = append(parts, controlLine)
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(green).Padding(1, 4).Align(lipgloss.Center).Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func helpLine(key, action string, rest ...string) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(gold)
	actionStyle := lipgloss.NewStyle().Foreground(muted)
	parts := []string{keyStyle.Render("[" + key + "]"), actionStyle.Render(" " + action)}
	for i := 0; i+1 < len(rest); i += 2 {
		parts = append(parts, "   ", keyStyle.Render("["+rest[i]+"]"), actionStyle.Render(" "+rest[i+1]))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

func renderHand(hand []card, reveal bool) string {
	if len(hand) == 0 {
		return ""
	}
	faces := make([]string, 0, len(hand))
	for i, c := range hand {
		if !reveal && i == 1 {
			faces = append(faces, cardBack.Render("┌─────┐\n│ ▒▒▒ │\n│ ▒▒▒ │\n│ ▒▒▒ │\n└─────┘"))
			continue
		}
		// Style complete rows separately: styling embedded fragments corrupts
		// visible-width calculations and can draw black artifacts in terminals.
		face := lipgloss.JoinVertical(lipgloss.Left,
			cardFace.Render("┌─────┐"),
			lipgloss.NewStyle().Width(7).Foreground(c.color()).Background(lipgloss.Color("252")).Bold(true).Render("│"+fmt.Sprintf("%-5s", c.rank)+"│"),
			lipgloss.NewStyle().Width(7).Foreground(c.color()).Background(lipgloss.Color("252")).Align(lipgloss.Center).Render("│  "+c.symbol()+"  │"),
			lipgloss.NewStyle().Width(7).Foreground(c.color()).Background(lipgloss.Color("252")).Bold(true).Render("│"+fmt.Sprintf("%5s", c.rank)+"│"),
			cardFace.Render("└─────┘"),
		)
		faces = append(faces, face)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, faces...)
}

func main() {
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "blackjack:", err)
		os.Exit(1)
	}
}
