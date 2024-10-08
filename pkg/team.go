package pkg

import "sync"

// Constants
const (
	kMaxApplicantSize         = 20
	kFiveMemberMaxSize        = 5
	kTenMemberMaxSize         = 10
	kMaxTeamSize              = 10000
	kInvalidGuid       uint64 = 0 // Assuming Guid is uint64
)

const (
	kOK                          = 0
	kTeamNotInApplicants         = 5000
	kTeamPlayerId                = 5001
	kTeamMembersFull             = 5002
	kTeamMemberInTeam            = 5003
	kTeamMemberNotInTeam         = 5004
	kTeamKickSelf                = 5005
	kTeamKickNotLeader           = 5006
	kTeamAppointSelf             = 5007
	kTeamAppointLeaderNotLeader  = 5008
	kTeamFull                    = 5009
	kTeamInApplicantList         = 5010
	kTeamNotInApplicantList      = 5011
	kTeamListMaxSize             = 5012
	kTeamHasNotTeamId            = 5013
	kTeamDismissNotLeader        = 5014
	kTeamJoinTeamMemberListToMax = 5015
	kTeamCreateTeamMaxMemberSize = 5016
	kTeamPlayerNotFound          = 5017
	kTeamApplyExist              = 5018
	kTeamAppointNotLeader        = 5019
	kTeamApplyJoin               = 5020
	kTeamApplyListFull           = 5021
)

// GuidVector is a slice of Guid (uint64)
type GuidVector []uint64

// CreateTeamParam represents parameters for creating a team
type CreateTeamParam struct {
	LeaderID     uint64
	MemberList   GuidVector
	TeamTypeSize uint64
}

// Team represents a team entity
type Team struct {
	LeaderID     uint64
	ID           uint64 // Assuming ID is uint64
	MemberList   GuidVector
	Applicants   GuidVector
	TeamTypeSize uint64
}

// TeamSystem represents the system managing teams
type TeamSystem struct {
	teams       map[uint64]*Team // Map of team ID to Team
	playerLists sync.Map         // Map of player ID to team ID
	lastTeamID  uint64           // For testing
}

func NewCreateTeamParam(leaderID uint64, members []uint64, teamTypeSize ...uint64) CreateTeamParam {
	// Default TeamTypeSize is 5
	size := uint64(5)
	if len(teamTypeSize) > 0 {
		size = teamTypeSize[0]
	}

	return CreateTeamParam{
		LeaderID:     leaderID,
		MemberList:   members,
		TeamTypeSize: size,
	}
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
		return len(team.MemberList)
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
		return len(team.MemberList) >= int(team.TeamTypeSize)
	}
	return false
}

func (ts *TeamSystem) HasMember(teamID, guid uint64) bool {
	if team, ok := ts.teams[teamID]; ok {
		for _, member := range team.MemberList {
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

func (ts *TeamSystem) CreateTeam(param CreateTeamParam) uint32 {
	// Check if the team list has reached its maximum size
	if ts.IsTeamListMax() {
		return kTeamListMaxSize
	}

	// Check if the leader is already in a team
	if ts.HasTeam(param.LeaderID) {
		return kTeamMemberInTeam
	}

	// Validate the number of members and check if all members are valid
	if len(param.MemberList) > int(param.TeamTypeSize) {
		return kTeamCreateTeamMaxMemberSize
	}
	if err := ts.CheckMemberInTeam(param.MemberList); err != kOK {
		return err
	}

	// Create a new team with a new ID
	teamID := ts.lastTeamID + 1
	ts.lastTeamID = teamID

	// Initialize the new team
	team := &Team{
		LeaderID:     param.LeaderID,
		ID:           teamID,
		MemberList:   make(GuidVector, len(param.MemberList)),
		Applicants:   make(GuidVector, 0),
		TeamTypeSize: param.TeamTypeSize,
	}
	copy(team.MemberList, param.MemberList)
	ts.teams[teamID] = team

	// Update player to team mappings
	for _, member := range param.MemberList {
		ts.playerLists.Store(member, teamID)
	}

	return kOK
}

func (ts *TeamSystem) JoinTeam(teamID, guid uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if ts.HasTeam(guid) {
			return kTeamMemberInTeam
		}
		if ts.IsTeamFull(teamID) {
			return kTeamMembersFull
		}
		if idx := ts.FindApplicantIndex(team, guid); idx != -1 {
			team.Applicants = append(team.Applicants[:idx], team.Applicants[idx+1:]...)
		}
		team.MemberList = append(team.MemberList, guid)
		ts.playerLists.Store(guid, teamID)
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) JoinTeamByMemberList(memberList GuidVector, teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if len(team.MemberList)+len(memberList) > int(team.TeamTypeSize) {
			return kTeamJoinTeamMemberListToMax
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
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) CheckMemberInTeam(memberList GuidVector) uint32 {
	for _, member := range memberList {
		if ts.HasTeam(member) {
			return kTeamMemberInTeam
		}
	}
	return kOK
}

func (ts *TeamSystem) LeaveTeam(guid uint64) uint32 {
	teamID := ts.GetTeamID(guid)
	if team, ok := ts.teams[teamID]; ok {
		if !ts.HasMember(teamID, guid) {
			return kTeamMemberNotInTeam
		}
		isLeaderLeave := team.LeaderID == guid
		ts.DelMember(teamID, guid)
		if len(team.MemberList) > 0 && isLeaderLeave {
			ts.OnAppointLeader(teamID, team.MemberList[0])
		}
		if len(team.MemberList) == 0 {
			ts.EraseTeam(teamID)
		}
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) KickMember(teamID, currentLeaderID, beKickID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID != currentLeaderID {
			return kTeamKickNotLeader
		}
		if team.LeaderID == beKickID || currentLeaderID == beKickID {
			return kTeamKickSelf
		}
		if !ts.HasMember(teamID, beKickID) {
			return kTeamMemberNotInTeam
		}
		ts.DelMember(teamID, beKickID)
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) Disbanded(teamID, currentLeaderID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID != currentLeaderID {
			return kTeamDismissNotLeader
		}
		for _, member := range team.MemberList {
			ts.DelMember(teamID, member)
		}
		ts.EraseTeam(teamID)
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) DisbandedTeamNoLeader(teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		return ts.Disbanded(teamID, team.LeaderID)
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) AppointLeader(teamID, currentLeaderID, newLeaderID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		if team.LeaderID == newLeaderID {
			return kTeamAppointSelf
		}
		if team.LeaderID != currentLeaderID {
			return kTeamAppointNotLeader
		}
		if !ts.HasMember(teamID, newLeaderID) {
			return kTeamMemberNotInTeam
		}
		ts.OnAppointLeader(teamID, newLeaderID)
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) ApplyToTeam(teamID, guid uint64) uint32 {
	team, ok := ts.teams[teamID]
	if !ok {
		// Team with teamID does not exist
		return kTeamHasNotTeamId
	}

	// Check if the user is already in a team
	if ts.HasTeam(guid) {
		return kTeamMemberInTeam
	}

	// Check if the user is already a member of the team
	if ts.HasMember(teamID, guid) {
		return kTeamApplyExist
	}

	if ts.IsTeamFull(teamID) {
		return kTeamMembersFull
	}

	// Check if the user is already an applicant
	if ts.IsApplicant(teamID, guid) {
		return kTeamApplyJoin
	}

	// If the applicants list is full, remove the oldest applicant
	if len(team.Applicants) >= kMaxApplicantSize {
		// Remove the first applicant from the list
		team.Applicants = team.Applicants[1:]
	}

	// Add the user to the applicant list
	team.Applicants = append(team.Applicants, guid)
	return kOK
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
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) ClearApplyList(teamID uint64) uint32 {
	if team, ok := ts.teams[teamID]; ok {
		team.Applicants = make(GuidVector, 0)
		return kOK
	}
	return kTeamHasNotTeamId
}

func (ts *TeamSystem) EraseTeam(teamID uint64) {
	if team, ok := ts.teams[teamID]; ok {
		for _, member := range team.MemberList {
			ts.playerLists.Delete(member)
		}
		delete(ts.teams, teamID)
	}
}

func (ts *TeamSystem) DelMember(teamID, guid uint64) {
	if team, ok := ts.teams[teamID]; ok {
		for idx, member := range team.MemberList {
			if member == guid {
				team.MemberList = append(team.MemberList[:idx], team.MemberList[idx+1:]...)
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
