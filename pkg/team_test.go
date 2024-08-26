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
func TestKickTeamMember(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)
	leaderPlayerID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	if got := ts.KickMember(ts.LastTeamID(), memberID, memberID); got != kTeamKickSelf {
		t.Errorf("KickMember() = %v, want %v", got, kTeamKickSelf)
	}
	if got := ts.KickMember(ts.LastTeamID(), 99, 99); got != kTeamKickNotLeader {
		t.Errorf("KickMember() = %v, want %v", got, kTeamKickNotLeader)
	}

	memberID++
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kOK {
		t.Errorf("JoinTeam() = %v, want %v", got, kOK)
	}
	if got := ts.KickMember(ts.LastTeamID(), leaderPlayerID, leaderPlayerID); got != kTeamKickSelf {
		t.Errorf("KickMember() = %v, want %v", got, kTeamKickSelf)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.KickMember(ts.LastTeamID(), memberID, leaderPlayerID); got != kTeamKickNotLeader {
		t.Errorf("KickMember() = %v, want %v", got, kTeamKickNotLeader)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.KickMember(ts.LastTeamID(), memberID, memberID); got != kTeamKickNotLeader {
		t.Errorf("KickMember() = %v, want %v", got, kTeamKickNotLeader)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.KickMember(ts.LastTeamID(), leaderPlayerID, 88); got != kTeamMemberNotInTeam {
		t.Errorf("KickMember() = %v, want %v", got, kTeamMemberNotInTeam)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.KickMember(ts.LastTeamID(), leaderPlayerID, memberID); got != kOK {
		t.Errorf("KickMember() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}

	if got := ts.TeamSize(); got != 1 {
		t.Errorf("GetTeamSize() = %v, want %v", got, 1)
	}
	if got := ts.PlayersSize(); got != 1 {
		t.Errorf("GetPlayersSize() = %v, want %v", got, 1)
	}
}

func TestAppointLeaderAndLeaveTeam1(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)
	leaderPlayerID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	playerID := memberID
	for i := 1; i < kFiveMemberMaxSize; i++ {
		memberID = playerID + uint64(i)
		if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kOK {
			t.Errorf("JoinTeam() = %v, want %v", got, kOK)
		}
		if got := ts.MemberSize(ts.LastTeamID()); got != i+1 {
			t.Errorf("MemberSize() = %v, want %v", got, i+1)
		}
	}

	if got := ts.AppointLeader(ts.LastTeamID(), leaderPlayerID, leaderPlayerID); got != kTeamAppointSelf {
		t.Errorf("AppointLeader() = %v, want %v", got, kTeamAppointSelf)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.AppointLeader(ts.LastTeamID(), 101, 100); got != kTeamAppointSelf {
		t.Errorf("AppointLeader() = %v, want %v", got, kTeamAppointSelf)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}
	if got := ts.AppointLeader(ts.LastTeamID(), 100, 100); got != kTeamAppointSelf {
		t.Errorf("AppointLeader() = %v, want %v", got, kTeamAppointSelf)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}

	if got := ts.AppointLeader(ts.LastTeamID(), 100, 101); got != kOK {
		t.Errorf("AppointLeader() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 101 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 101)
	}

	ts.LeaveTeam(101)

	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}

	leaderPlayerID += 2
	if got := ts.AppointLeader(ts.LastTeamID(), 100, 102); got != kOK {
		t.Errorf("AppointLeader() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}

	ts.LeaveTeam(102)
	leaderPlayerID = 100
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != leaderPlayerID {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, leaderPlayerID)
	}

	if got := ts.AppointLeader(ts.LastTeamID(), 100, 103); got != kOK {
		t.Errorf("AppointLeader() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 103 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 103)
	}

	ts.LeaveTeam(103)
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 100 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 100)
	}

	if got := ts.AppointLeader(ts.LastTeamID(), 100, 104); got != kOK {
		t.Errorf("AppointLeader() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 104 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 104)
	}

	ts.LeaveTeam(104)
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 100 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 100)
	}

	ts.LeaveTeam(100)
	if ts.HasTeam(100) {
		t.Errorf("Expected team with ID 100 to be removed")
	}
}

func TestAppointLeaderAndLeaveTeam2(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	memberID = 104
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kOK {
		t.Errorf("JoinTeam() = %v, want %v", got, kOK)
	}

	if got := ts.AppointLeader(ts.LastTeamID(), 100, 104); got != kOK {
		t.Errorf("AppointLeader() = %v, want %v", got, kOK)
	}
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 104 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 104)
	}

	ts.LeaveTeam(100)
	if got := ts.GetLeaderIDByTeamID(ts.LastTeamID()); got != 104 {
		t.Errorf("GetLeaderIDByTeamID() = %v, want %v", got, 104)
	}

	ts.LeaveTeam(104)
	if ts.HasTeam(104) {
		t.Errorf("Expected team with ID 104 to be removed")
	}
}

func TestDismissTeam(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(100)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	memberID = 104
	if got := ts.JoinTeam(ts.LastTeamID(), memberID); got != kOK {
		t.Errorf("JoinTeam() = %v, want %v", got, kOK)
	}

	if got := ts.Disbanded(ts.LastTeamID(), 104); got != kTeamDismissNotLeader {
		t.Errorf("Disbanded() = %v, want %v", got, kTeamDismissNotLeader)
	}
	if got := ts.Disbanded(111, 104); got != kTeamHasNotTeamId {
		t.Errorf("Disbanded() = %v, want %v", got, kTeamHasNotTeamId)
	}
	if got := ts.Disbanded(ts.LastTeamID(), 100); got != kOK {
		t.Errorf("Disbanded() = %v, want %v", got, kOK)
	}
	if ts.HasTeam(100) {
		t.Errorf("Expected team with ID 100 to be removed")
	}
}

func TestApplyFull(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(1001)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	nMax := uint64(kMaxApplicantSize * 2)
	for i := uint64(0); i < nMax; i++ {
		app := i
		if got := ts.ApplyToTeam(ts.LastTeamID(), app); got != kOK {
			t.Errorf("ApplyToTeam() = %v, want %v", got, kOK)
		}
		if i < kMaxApplicantSize {
			if got := ts.ApplicantSizeByTeamID(ts.LastTeamID()); got != int(i+1) {
				t.Errorf("GetApplicantSizeByTeamID() = %v, want %v", got, i+1)
			}
		} else {
			if got := ts.ApplicantSizeByTeamID(ts.LastTeamID()); got != kMaxApplicantSize {
				t.Errorf("GetApplicantSizeByTeamID() = %v, want %v", got, kMaxApplicantSize)
			}
			if got := ts.FirstApplicant(ts.LastTeamID()); got != i-kMaxApplicantSize+1 {
				t.Errorf("GetFirstApplicant() = %v, want %v", got, i-kMaxApplicantSize+1)
			}
		}
	}

	for i := uint64(0); i < uint64(nMax-kMaxApplicantSize); i++ {
		if ts.IsApplicant(ts.LastTeamID(), i) {
			t.Errorf("Expected applicant %v to be not in the team", i)
		}
	}

	for i := nMax - 10; i < nMax; i++ {
		if !ts.IsApplicant(ts.LastTeamID(), i) {
			t.Errorf("Expected applicant %v to be in the team", i)
		}
	}
}

func TestApplicantOrder(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(1001)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	nMax := uint64(kMaxApplicantSize)
	for i := uint64(0); i < nMax; i++ {
		a := i
		if got := ts.ApplyToTeam(ts.LastTeamID(), a); got != kOK {
			t.Errorf("ApplyToTeam() = %v, want %v", got, kOK)
		}
	}
	if got := ts.FirstApplicant(ts.LastTeamID()); got != nMax-kMaxApplicantSize {
		t.Errorf("GetFirstApplicant() = %v, want %v", got, nMax-kMaxApplicantSize)
	}

	secondMax := nMax
	for i := secondMax; i < secondMax+nMax; i++ {
		a := i
		if got := ts.ApplyToTeam(ts.LastTeamID(), a); got != kOK {
			t.Errorf("ApplyToTeam() = %v, want %v", got, kOK)
		}
	}

	if got := ts.FirstApplicant(ts.LastTeamID()); got != secondMax {
		t.Errorf("GetFirstApplicant() = %v, want %v", got, secondMax)
	}
}

func TestInTeamApplyForTeam(t *testing.T) {
	ts := NewTeamSystem()
	memberID := uint64(1001)

	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}

	nMax := uint64(kMaxApplicantSize)
	for i := uint64(1); i < nMax; i++ {
		a := i
		if got := ts.ApplyToTeam(ts.LastTeamID(), a); got != kOK {
			t.Errorf("ApplyToTeam() = %v, want %v", got, kOK)
		}
	}
	for i := uint64(1); i < nMax; i++ {
		if i < kFiveMemberMaxSize {
			if got := ts.JoinTeam(ts.LastTeamID(), i); got != kOK {
				t.Errorf("JoinTeam() = %v, want %v", got, kOK)
			}
			if ts.IsApplicant(ts.LastTeamID(), i) {
				t.Errorf("Expected applicant %v to be not in the team", i)
			}
		} else {
			if got := ts.JoinTeam(ts.LastTeamID(), i); got != kTeamMembersFull {
				t.Errorf("JoinTeam() = %v, want %v", got, kTeamMembersFull)
			}
			if !ts.IsApplicant(ts.LastTeamID(), i) {
				t.Errorf("Expected applicant %v to be in the team", i)
			}
		}
	}

	a := uint64(6666)
	if got := ts.ApplyToTeam(ts.LastTeamID(), a); got != kTeamMembersFull {
		t.Errorf("ApplyToTeam() = %v, want %v", got, kTeamMembersFull)
	}

	if got := ts.LeaveTeam(2); got != kOK {
		t.Errorf("LeaveTeam() = %v, want %v", got, kOK)
	}

	memberID = 2
	if got := ts.ApplyToTeam(ts.LastTeamID(), memberID); got != kOK {
		t.Errorf("ApplyToTeam() = %v, want %v", got, kOK)
	}
	if got := ts.CreateTeam(NewCreateTeamParam(memberID, []uint64{memberID})); got != kOK {
		t.Errorf("CreateTeam() = %v, want %v", got, kOK)
	}
	if got := ts.JoinTeam(ts.LastTeamID(), 2); got != kTeamMemberInTeam {
		t.Errorf("JoinTeam() = %v, want %v", got, kTeamMemberInTeam)
	}
	if ts.IsApplicant(ts.LastTeamID(), 2) {
		t.Errorf("Expected applicant 2 to be not in the team")
	}
}
