package main

import (
	"fmt"
	"sync"
)

// Constants
const (
	kMaxApplicantSize         = 20
	kFiveMemberMaxSize        = 5
	kTenMemberMaxSize         = 10
	kMaxTeamSize              = 10000
	kInvalidGuid       uint64 = 0 // Assuming Guid is uint64
)

const (
	kOK                             = 0
	kRetTeamNotInApplicants         = 5000
	kRetTeamPlayerId                = 5001
	kRetTeamMembersFull             = 5002
	kRetTeamMemberInTeam            = 5003
	kRetTeamMemberNotInTeam         = 5004
	kRetTeamKickSelf                = 5005
	kRetTeamKickNotLeader           = 5006
	kRetTeamAppointSelf             = 5007
	kRetTeamAppointLeaderNotLeader  = 5008
	kRetTeamFull                    = 5009
	kRetTeamInApplicantList         = 5010
	kRetTeamNotInApplicantList      = 5011
	kRetTeamListMaxSize             = 5012
	kRetTeamHasNotTeamId            = 5013
	kRetTeamDismissNotLeader        = 5014
	kRetTeamJoinTeamMemberListToMax = 5015
	kRetTeamCreateTeamMaxMemberSize = 5016
	kRetTeamPlayerNotFound          = 5017
	kRetTeamApplyExist              = 5018
	kRetTeamAppointNotLeader        = 5019
	kRetTeamApplyJoin               = 5020
	kRetTeamApplyListFull           = 5021
)

// GuidVector is a slice of Guid (uint64)
type GuidVector []uint64

// CreateTeamP represents parameters for creating a team
type CreateTeamP struct {
	LeaderID     uint64
	MemberList   GuidVector
	TeamTypeSize uint64
}

// Team represents a team entity
type Team struct {
	LeaderID     uint64
	TeamID       uint64 // Assuming TeamID is uint64
	Members      GuidVector
	Applicants   GuidVector
	TeamTypeSize uint64
}

// TeamSystem represents the system managing teams
type TeamSystem struct {
	teams       map[uint64]*Team // Map of team ID to Team
	playerLists sync.Map         // Map of player ID to team ID
	lastTeamID  uint64           // For testing
}

// NewTeamSystem initializes a new TeamSystem
func NewTeamSystem() *TeamSystem {
	return &TeamSystem{
		teams: make(map[uint64]*Team),
	}
}

// Methods of TeamSystem

func (ts *TeamSystem) TeamSize() int {
	return len(ts.teams)
}

func (ts *TeamSystem) LastTeamID() uint64 {
	return ts.lastTeamID
}

func (ts *TeamSystem) IsTeamListMax() bool {
	return len(ts.teams) >= kMaxTeamSize
}

func (ts *TeamSystem) MemberSize(teamID uint64) int {
	if team, ok := ts.teams[teamID]; ok {
		return len(team.Members)
	}
	return 0
}

func (ts *TeamSystem) ApplicantSizeByPlayerID(guid uint64) int {
	teamID := ts.GetTeamID(guid)
	return ts.ApplicantSizeByTeamID(teamID)
}

func (ts *TeamSystem) ApplicantSizeByTeamID(teamID uint64) int {
	if team, ok := ts.teams[teamID]; ok {
		return len(team.Applicants)
	}
	return 0
}

func (ts *TeamSystem) PlayersSize() int {
	count := 0
	ts.playerLists.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (ts *TeamSystem) GetTeamID(guid uint64) uint64 {
	if teamID, ok := ts.playerLists.Load(guid); ok {
		return teamID.(uint64)
	}
	return kInvalidGuid
}

func (ts *TeamSystem) GetLeaderIDByTeamID(teamID uint64) uint64 {
	if team, ok := ts.teams[teamID]; ok {
		return team.LeaderID
	}
	return kInvalidGuid
}

func (ts *TeamSystem) GetLeaderIDByPlayerID(guid uint64) uint64 {
	teamID := ts.GetTeamID(guid)
	return ts.GetLeaderIDByTeamID(teamID)
}

func (ts *TeamSystem) FirstApplicant(teamID uint64) uint64 {
	if team, ok := ts.teams[teamID]; ok && len(team.Applicants) > 0 {
		return team.Applicants[0]
	}
	return kInvalidGuid
}

func (ts *TeamSystem) IsTeamFull(teamID uint64) bool {
	if team, ok := ts.teams[teamID]; ok {
		return len(team.Members) >= int(team.TeamTypeSize)
	}
	return false
}

func (ts *TeamSystem) HasMember(teamID, guid uint64) bool {
	if team, ok := ts.teams[teamID]; ok {
		for _, member := range team.Members {
			if member == guid {
				return true
			}
		}
	}
	return false
}

func (ts *TeamSystem) HasTeam(guid uint64) bool {
	_, ok := ts.playerLists.Load(guid)
	return ok
}

func (ts *TeamSystem) IsApplicant(teamID, guid uint64) bool {
	if team, ok := ts.teams[teamID]; ok {
		for _, applicant := range team.Applicants {
			if applicant == guid {
				return true
			}
		}
	}
	return false
}

func (ts *TeamSystem) CreateTeam(param CreateTeamP) uint32 {
	if ts.IsTeamListMax() {
		return kRetTeamListMaxSize
	}
	if ts.HasTeam(param.LeaderID) {
		return kRetTeamMemberInTeam
	}
	if len(param.MemberList) > int(param.TeamTypeSize) {
		return kRetTeamCreateTeamMaxMemberSize
	}
	if err := ts.CheckMemberInTeam(param.MemberList); err != kOK {
		return err
	}
	teamID := ts.lastTeamID + 1
	ts.lastTeamID = teamID

	team := &Team{
		LeaderID:     param.LeaderID,
		TeamID:       teamID,
		Members:      make(GuidVector, len(param.MemberList)),
		Applicants:   make(GuidVector, 0),
		TeamTypeSize: param.TeamTypeSize,
	}
	copy(team.Members, param.MemberList)
	ts.teams[teamID] = team

	for _, member := range param.MemberList {
		ts.playerLists.Store(member, teamID)
	}

	return kOK
}

func (ts *TeamSystem) JoinTeam(teamID, guid uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if ts.HasTeam(guid) {
			return kRetTeamMemberInTeam
		}
		if ts.IsTeamFull(teamID) {
			return kRetTeamMembersFull
		}
		if idx := ts.FindApplicantIndex(team, guid); idx != -1 {
			team.Applicants = append(team.Applicants[:idx], team.Applicants[idx+1:]...)
		}
		team.Members = append(team.Members, guid)
		ts.playerLists.Store(guid, teamID)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) JoinTeamByMemberList(memberList GuidVector, teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if len(team.Members)+len(memberList) > int(team.TeamTypeSize) {
			return kRetTeamJoinTeamMemberListToMax
		}
		if err := ts.CheckMemberInTeam(memberList); err != kOK {
			return err
		}
		for _, member := range memberList {
			if err := ts.JoinTeam(teamID, member); err != kOK {
				return err
			}
		}
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) CheckMemberInTeam(memberList GuidVector) uint32 {
	for _, member := range memberList {
		if ts.HasTeam(member) {
			return kRetTeamMemberInTeam
		}
	}
	return kOK
}

func (ts *TeamSystem) LeaveTeam(guid uint64) uint32 {
	teamID := ts.GetTeamID(guid)
	if team, ok := ts.teams[teamID]; ok {
		if !ts.HasMember(teamID, guid) {
			return kRetTeamMemberNotInTeam
		}
		isLeaderLeave := team.LeaderID == guid
		ts.DelMember(teamID, guid)
		if len(team.Members) > 0 && isLeaderLeave {
			ts.OnAppointLeader(teamID, team.Members[0])
		}
		if len(team.Members) == 0 {
			ts.EraseTeam(teamID)
		}
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) KickMember(teamID, currentLeaderID, beKickID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID != currentLeaderID {
			return kRetTeamKickNotLeader
		}
		if team.LeaderID == beKickID || currentLeaderID == beKickID {
			return kRetTeamKickSelf
		}
		if !ts.HasMember(teamID, beKickID) {
			return kRetTeamMemberNotInTeam
		}
		ts.DelMember(teamID, beKickID)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) Disbanded(teamID, currentLeaderID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID != currentLeaderID {
			return kRetTeamDismissNotLeader
		}
		for _, member := range team.Members {
			ts.DelMember(teamID, member)
		}
		ts.EraseTeam(teamID)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) DisbandedTeamNoLeader(teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		return ts.Disbanded(teamID, team.LeaderID)
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) AppointLeader(teamID, currentLeaderID, newLeaderID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID != currentLeaderID {
			return kRetTeamAppointNotLeader
		}
		if !ts.HasMember(teamID, newLeaderID) {
			return kRetTeamMemberNotInTeam
		}
		ts.OnAppointLeader(teamID, newLeaderID)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) ApplyToTeam(teamID, guid uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if ts.HasTeam(guid) {
			return kRetTeamMemberInTeam
		}
		if ts.HasMember(teamID, guid) {
			return kRetTeamApplyExist
		}
		if ts.IsApplicant(teamID, guid) {
			return kRetTeamApplyJoin
		}
		if len(team.Applicants) >= kMaxApplicantSize {
			return kRetTeamApplyListFull
		}
		team.Applicants = append(team.Applicants, guid)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) DelApplicant(teamID, guid uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		for idx, applicant := range team.Applicants {
			if applicant == guid {
				team.Applicants = append(team.Applicants[:idx], team.Applicants[idx+1:]...)
				return kOK
			}
		}
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) ClearApplyList(teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		team.Applicants = make(GuidVector, 0)
		return kOK
	}
	return kRetTeamHasNotTeamId
}

func (ts *TeamSystem) EraseTeam(teamID uint64) {
	if team, ok := ts.teams[teamID]; ok {
		for _, member := range team.Members {
			ts.playerLists.Delete(member)
		}
		delete(ts.teams, teamID)
	}
}

func (ts *TeamSystem) DelMember(teamID, guid uint64) {
	if team, ok := ts.teams[teamID]; ok {
		for idx, member := range team.Members {
			if member == guid {
				team.Members = append(team.Members[:idx], team.Members[idx+1:]...)
				ts.playerLists.Delete(guid)
				return
			}
		}
	}
}

func (ts *TeamSystem) OnAppointLeader(teamID, newLeaderID uint64) {
	if team, ok := ts.teams[teamID]; ok {
		team.LeaderID = newLeaderID
	}
}

func (ts *TeamSystem) FindApplicantIndex(team *Team, guid uint64) int {
	for idx, applicant := range team.Applicants {
		if applicant == guid {
			return idx
		}
	}
	return -1
}

func main() {
	// Example usage
	ts := NewTeamSystem()

	param := CreateTeamP{
		LeaderID:     1,
		MemberList:   GuidVector{2, 3, 4},
		TeamTypeSize: 5,
	}

	result := ts.CreateTeam(param)
	fmt.Println("Create team result:", result)

	fmt.Println("Team size:", ts.TeamSize())
	fmt.Println("Last team ID:", ts.LastTeamID())

	// Other operations can be tested similarly
}
