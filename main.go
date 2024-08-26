package main

import (
	"fmt"
	"team/pkg"
)

func main() {
	// Example usage
	ts := pkg.NewTeamSystem()

	param := pkg.CreateTeamParam{
		LeaderID:     1,
		MemberList:   pkg.GuidVector{2, 3, 4},
		TeamTypeSize: 5,
	}

	result := ts.CreateTeam(param)
	fmt.Println("Create team result:", result)

	fmt.Println("Team size:", ts.TeamSize())
	fmt.Println("Last team ID:", ts.LastTeamID())

	// Other operations can be tested similarly
}
