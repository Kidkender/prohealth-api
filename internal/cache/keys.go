package cache

import "fmt"

func MemberRoleKey(hospitalID, userID uint) string {
	return fmt.Sprintf("member:%d:%d", hospitalID, userID)
}

func ScheduleKey(hospitalID uint) string {
	return fmt.Sprintf("schedule:%d", hospitalID)
}

func ScheduleAllKey() string {
	return "schedule:all"
}
