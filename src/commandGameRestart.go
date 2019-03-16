/*
	Sent when the user is in a shared replay of a speedrun game
	and wants to start a new game with the same settings as the current game
	"data" is empty
*/

package main

import (
	"strconv"
)

func commandGameRestart(s *Session, d *CommandData) {
	/*
		Validate
	*/

	// Validate that the game exists
	gameID := s.CurrentGame()
	var g *Game
	if v, ok := games[gameID]; !ok {
		s.Warning("Game " + strconv.Itoa(gameID) + " does not exist.")
		return
	} else {
		g = v
	}

	// Validate that this is a shared replay
	if !g.Replay || !g.Visible {
		s.Warning("Game " + strconv.Itoa(gameID) + " is not a shared replay, so you cannot send a restart action.")
		return
	}

	// Validate that this person is leading the shared replay
	if s.UserID() != g.Owner {
		s.Warning("You cannot restart a game unless you are the leader.")
		return
	}

	// Validate that there are at least two people in the shared replay
	if len(g.Spectators) < 2 {
		s.Warning("You cannot restart a game unless there are at least two people in it.")
		return
	}

	// Validate that all of the players who played the game are currently spectating
	// the shared replay
	gameSessions := make([]*Session, 0)
	for _, sp := range g.Spectators {
		playedInOriginalGame := false
		for _, p := range g.Players {
			if p.Name == sp.Name {
				playedInOriginalGame = true
				break
			}
		}
		if playedInOriginalGame {
			gameSessions = append(gameSessions, sp.Session)
		}
	}
	if len(gameSessions) != len(g.Players) {
		s.Warning("Not all of the players from the original game are in the shared replay, " +
			"so you cannot restart the game.")
		return
	}

	/*
		Restart
	*/

	// Force the client of all of the spectators to go back to the lobby
	g.NotifyBoot()

	// On the server side, all of the spectators will still be in the game,
	// so manually disconnect them
	for _, s2 := range gameSessions {
		commandGameUnattend(s2, nil)
	}

	// The shared replay should now be deleted, since all of the players have left
	// Now, emulate the players creating and joining and starting a new game
	commandGameCreate(s, &CommandData{
		Name:                 getName(), // Generate a random name for the new game
		Variant:              g.Options.Variant,
		Timed:                g.Options.Timed,
		BaseTime:             g.Options.BaseTime,
		TimePerTurn:          g.Options.TimePerTurn,
		Speedrun:             g.Options.Speedrun,
		DeckPlays:            g.Options.DeckPlays,
		EmptyClues:           g.Options.EmptyClues,
		CharacterAssignments: g.Options.CharacterAssignments,
	})
	for _, s2 := range gameSessions {
		if s2.UserID() == s.UserID() {
			// The creator of the game does not need to join
			continue
		}
		commandGameJoin(s2, &CommandData{
			// We increment the newGameID after creating a game,
			// so assume that the ID of the last game created is equal to the "newGameID" minus 1
			ID: newGameID - 1,
		})
	}
	commandGameStart(s, nil)
}
