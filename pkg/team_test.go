package pkg

import (
	"testing"
)

// Helper function to create a team and return its ID
func createTeam(ts *TeamSystem, playerID uint64) uint64 {
	params := NewCreateTeamParam(playerID, []uint64{playerID})
	ts.CreateTeam(params)
	return ts.LastTeamID()
}

func TestCreateFullDismiss(t *testing.T) {
	ts := NewTeamSystem()
	teamIDs := make([]uint64, 0, kMaxTeamSize)
	playerID := uint64(1)

	for i := 0; i < kMaxTeamSize; i++ {
		teamID := createTeam(ts, playerID)
		teamIDs = append(teamIDs, teamID)
		playerID++
	}

	if !ts.IsTeamListMax() {
		t.Errorf("Expected team list to be at max size")
	}

	playerID++
	if got := ts.CreateTeam(CreateTeamParam{LeaderID: playerID, MemberList: []uint64{playerID}}); got != kTeamListMaxSize {
		t.Errorf("CreateTeam() = %v, want %v", got, kTeamListMaxSize)
	}

	if got := ts.TeamSize(); got != kMaxTeamSize {
		t.Errorf("GetTeamSize() = %v, want %v", got, kMaxTeamSize)
	}

	for _, teamID := range teamIDs {
		leaderID := ts.GetLeaderIDByTeamID(teamID)
		if got := ts.Disbanded(teamID, leaderID); got != kOK {
			t.Errorf("DisbandTeam() = %v, want %v", got, kOK)
		}
	}

	if got := ts.TeamSize(); got != 0 {
		t.Errorf("GetTeamSize() = %v, want %v", got, 0)
	}
	if got := ts.PlayersSize(); got != 0 {
		t.Errorf("GetPlayersSize() = %v, want %v", got, 0)
	}
}

func TestTeamSize(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}
	if !ts.HasMember(ts.LastTeamID(), memberID) {
		t.Errorf("Expected memberID to be in team")
	}
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kTeamMemberInTeam {
		t.Errorf("JoinTeam() = %v, want %v", got, kTeamMemberInTeam)
	}
	if got := ts.MemberSize(ts.LastTeamID()); got != 1 {
		t.Errorf("MemberSize() = %v, want %v", got, 1)
	}

	for i := 1; i < kFiveMemberMaxSize; i++ {
		memberID++
		if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kOK {
			t.Errorf("JoinTeam() = %v, want %v", got, kOK)
		}
		if got := ts.MemberSize(ts.LastTeamID()); got != i+1 {
			t.Errorf("MemberSize() = %v, want %v", got, i+1)
		}
	}

	memberID++
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kTeamMembersFull {
		t.Errorf("JoinTeam() = %v, want %v", got, kTeamMembersFull)
	}
	if got := ts.MemberSize(ts.LastTeamID()); got != kFiveMemberMaxSize {
		t.Errorf("MemberSize() = %v, want %v", got, kFiveMemberMaxSize)
	}
}

func TestLeaveTeam(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}
	if !ts.HasMember(ts.LastTeamID(), memberID) {
		t.Errorf("Expected memberID to be in team")
	}
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kTeamMemberInTeam {
		t.Errorf("JoinTeam() = %v, want %v", got, kTeamMemberInTeam)
	}
	if got := ts.MemberSize(ts.LastTeamID()); got != 1 {
		t.Errorf("MemberSize() = %v, want %v", got, 1)
	}

	ts.LeaveTeam(memberID)
	if ts.HasMember(ts.LastTeamID(), memberID) {
		t.Errorf("Expected memberID to not be in team")
	}
	if got := ts.MemberSize(ts.LastTeamID()); got != 0 {
		t.Errorf("MemberSize() = %v, want %v", got, 0)
	}
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kTeamHasNotTeamId {
		t.Errorf("JoinTeam() = %v, want %v", got, kTeamHasNotTeamId)
	}
	if got := ts.MemberSize(ts.LastTeamID()); got != 0 {
		t.Errorf("MemberSize() = %v, want %v", got, 0)
	}

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	playerID := memberID
	for i := uint64(1); i < kFiveMemberMaxSize; i++ {
		playerID += i
		if got := ts.JoinTeam(ts.LastTeamID(), playerID); got != kOK {
			t.Errorf("JoinTeam() = %v, want %v", got, kOK)
		}
		if got := ts.MemberSize(ts.LastTeamID()); uint64(got) != i+1 {
			t.Errorf("MemberSize() = %v, want %v", got, i+1)
		}
	}

	playerID = memberID
	for i := uint64(0); i < kFiveMemberMaxSize; i++ {
		playerID += i
		ts.LeaveTeam(playerID)
		if ts.HasMember(ts.LastTeamID(), playerID) {
			t.Errorf("Expected playerID to not be in team")
		}
		if i < 4 {
			if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != playerID+i+1 {
				t.Errorf("GetLeaderIdByTeamId() = %v, want %v", got, playerID+i+1)
			}
			if got := ts.MemberSize(ts.LastTeamID()); got != kFiveMemberMaxSize-int(i)-1 {
				t.Errorf("MemberSize() = %v, want %v", got, kFiveMemberMaxSize-i-1)
			}
		}
		if got := ts.MemberSize(ts.LastTeamID()); got != kFiveMemberMaxSize-int(i)-1 {
			t.Errorf("MemberSize() = %v, want %v", got, kFiveMemberMaxSize-i-1)
		}
	}
	if got := ts.TeamSize(); got != 0 {
		t.Errorf("GetTeamSize() = %v, want %v", got, 0)
	}
	if got := ts.PlayersSize(); got != 0 {
		t.Errorf("GetPlayersSize() = %v, want %v", got, 0)
	}
}
