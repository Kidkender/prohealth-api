package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kidkender/prohealth-api/internal/apperror"
	"github.com/kidkender/prohealth-api/internal/cache"
	"github.com/kidkender/prohealth-api/internal/dto"
	"github.com/kidkender/prohealth-api/internal/llm"
	"github.com/kidkender/prohealth-api/internal/model"
	"github.com/kidkender/prohealth-api/internal/security"
)

const maxToolIterations = 5
const sessionTTL = 30 * time.Minute

type ChatReply struct {
	SessionID string `json:"session_id"`
	Reply     string `json:"reply"`
}

type KnownPatientInfo struct {
	Fullname string
	Phone    string
	Email    string
}

type AgentService struct {
	cache              *cache.Cache
	llm                llm.LLM
	scheduleService    *ScheduleService
	hospitalService    *HospitalService
	departmentService  *DepartmentService
	appointmentService *AppointmentService
}

func NewAgentService(
	cache *cache.Cache,
	llm llm.LLM,
	scheduleService *ScheduleService,
	hospitalService *HospitalService,
	departmentService *DepartmentService,
	appointmentService *AppointmentService,
) *AgentService {
	return &AgentService{
		cache:              cache,
		llm:                llm,
		scheduleService:    scheduleService,
		hospitalService:    hospitalService,
		departmentService:  departmentService,
		appointmentService: appointmentService,
	}
}

func buildSystemPrompt(hospitalName string, departments []model.Department, knownPatient *KnownPatientInfo) llm.Message {
	now := time.Now()

	var deptList strings.Builder
	for _, d := range departments {
		fmt.Fprintf(&deptList, "- department_id=%d: %s\n", d.ID, d.Name)
	}

	content := fmt.Sprintf(`You are the appointment booking assistant for %s Hospital.

Available departments (you MUST use the exact department_id below when calling
tools — never invent or guess a department_id. If the user's requested
specialty does not clearly match any department in this list, ask the user to
choose one instead of guessing):

%s

Today's date is %s (%s).

Always resolve relative or partial date references against today's date.
Examples:
- "ngày mai"
- "thứ Hai"
- "tuần sau"
- "ngày 13"

Never invent a different month or year.

If the user only specifies a day of month (for example "ngày 13"), assume the
nearest upcoming occurrence starting from today.

Whenever you resolve an ambiguous date, always tell the user the exact resolved
date (dd/MM/yyyy) in your next reply.

Your role is to help users book medical appointments through a natural
conversation.

Booking workflow:

STEP 1 - Understand the request

Before calling "search_available_slots", you ONLY need:

- Department (or specialty)
- Preferred appointment date

If either of them is missing, ask only for the missing information.

The patient's full name, phone number, and email are NOT required to search
available appointments.

STEP 2 - Search

As soon as both department and date are known, immediately call
"search_available_slots".

Never delay the search because patient information has not yet been collected.

STEP 3 - Propose

If available appointments are found:

- Call "propose_slot" with BOTH "shift_id" and "start_time" (the exact HH:MM
  of the slot you are recommending, taken from the "search_available_slots"
  result — never omit or guess this value).
- Recommend ONE suitable slot.
- Tell the user the exact appointment date and time.
- Ask whether they would like to book it.

If no appointment is available:

- Inform the user politely.
- Suggest another date or another department.

STEP 4 - Collect patient information

Only AFTER the user accepts a proposed appointment slot, collect any missing
information:

- Full name
- Phone number
- Email (optional)

If the user is already authenticated and the information is already known,
NEVER ask again.

STEP 5 - Confirmation

After all required information has been collected, explicitly ask the user to
confirm the booking.

Only clear confirmations such as:

- "đồng ý"
- "đặt luôn"
- "ok"
- "yes"
- "confirm"

allow you to proceed.

STEP 6 - Create appointment

Only after ALL of the following conditions are satisfied:

- An appointment slot has been proposed.
- The user has accepted that slot.
- Required patient information is available.

may you call "create_appointment".

General rules:

1. Never invent doctors, schedules, appointment slots, or availability.

2. Tool outputs are the single source of truth.

3. Never guess a department_id.
Always use one of the department IDs listed above.

4. If the user changes the department or date after a slot has already been
proposed, discard the previous proposal and perform a new search.

5. Ask only for information that is still missing.

6. Never ask the user to repeat information that is already known.

7. Always respond in Vietnamese.

8. Keep your replies concise, natural, friendly, and professional, like a
hospital receptionist.
`, hospitalName, deptList.String(), now.Format("2006-01-02"), now.Weekday())

	if knownPatient != nil {
		content += fmt.Sprintf(`

The user is already authenticated.

The following information is already known and MUST NOT be requested again:

- Full name: %s
- Phone number: %s
`, knownPatient.Fullname, knownPatient.Phone)

		if knownPatient.Email != "" {
			content += fmt.Sprintf("- Email: %s\n", knownPatient.Email)
		}
	}

	return llm.Message{
		Role:    llm.RoleSystem,
		Content: content,
	}
}

// func buildSystemPrompt(hospitalName string, departments []model.Department, knownPatient *KnownPatientInfo) llm.Message {
// 	now := time.Now()

// 	var deptList strings.Builder
// 	for _, d := range departments {
// 		fmt.Fprintf(&deptList, "- department_id=%d: %s\n", d.ID, d.Name)
// 	}

// 	content := fmt.Sprintf(`You are the appointment booking assistant for %s Hospital.

// Available departments (you MUST use the exact department_id below when calling
// tools — never invent or guess a department_id; if the user's requested
// specialty doesn't clearly match any department in this list, ask them to pick
// from this list instead of guessing):
// %s

// Today's date is %s (%s). Always resolve relative or partial date references
// (e.g. "ngày 13", "ngày mai", "tuần sau") against this date — never assume or
// invent a different year or month. If the user gives a day-of-month without a
// month/year (e.g. "ngày 13"), assume the nearest upcoming occurrence of that
// day starting from today. Whenever you silently resolve an ambiguous date like
// this, you MUST state the exact resolved date (dd/mm/yyyy) back to the user in
// your very next reply, even if you are also asking for other missing
// information in the same message — never resolve a date without confirming it
// out loud.

// Your role is to help users book medical appointments through a conversational chat.

// Before searching for available appointment slots, you must collect the following required information:
// - Department or specialty the user wants to visit (ask if not specified)
// - Preferred appointment date
// - Patient's full name and phone number (email is optional)

// Mandatory rules:
// 1. Never invent or assume available doctors or appointment slots. Always call the "search_available_slots" tool to retrieve real availability before suggesting any appointment time.
// 2. After finding suitable availability, only call the "propose_slot" tool to present the option to the user. Do NOT call "create_appointment" immediately.
// 3. After proposing an appointment, explicitly ask the user to confirm the booking. Only call the "create_appointment" tool when the user clearly confirms (e.g. "yes", "confirm", "book it", "okay", etc.).
// 4. If the user changes their mind or asks for another option after a proposal has been made, perform a new search instead of reusing the previous proposal.
// 5. Always respond in Vietnamese. Keep responses concise, friendly, and natural, like a hospital receptionist.
// `, hospitalName, deptList.String(), now.Format("2006-01-02"), now.Weekday())

// 	if knownPatient != nil {
// 		content += fmt.Sprintf(`

// The user is already authenticated. The following information is already known and MUST NOT be requested again:
// - Full name: %s
// - Phone number: %s
// `, knownPatient.Fullname, knownPatient.Phone)
// 	}

// 	slog.Info("system prompt built", "content", content)

// 	return llm.Message{
// 		Role:    llm.RoleSystem,
// 		Content: content,
// 	}
// }

func (a *AgentService) toolSpecs() []llm.ToolSpec {
	return []llm.ToolSpec{
		{
			Name:        "search_available_slots",
			Description: "Tìm lịch khám còn trống theo khoa và ngày",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"department_id": {"type": "integer"},
					"date": {"type": "string", "description": "YYYY-MM-DD"}
				},
				"required": ["date"]
			}`),
		},
		{
			Name:        "propose_slot",
			Description: "Đề xuất 1 slot cụ thể cho người dùng xác nhận, chưa đặt lịch thật",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"shift_id": {"type": "integer"},
					"start_time": {"type": "string", "description": "HH:MM của slot được chọn, lấy từ kết quả search_available_slots"}
				},
				"required": ["shift_id", "start_time"]
			}`),
		},
		{
			Name:        "create_appointment",
			Description: "Chốt đặt lịch — chỉ gọi sau khi người dùng đã xác nhận rõ ràng slot đã đề xuất",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		},
	}
}

func (a *AgentService) HandleMessage(
	ctx context.Context,
	hospitalID uint,
	sessionID *string,
	userID *uint,
	message string,
) (*ChatReply, error) {
	if sessionID == nil {
		newID, err := security.GenerateRandomToken()
		if err != nil {
			return nil, err
		}
		sessionID = &newID
	}

	sessionKey := cache.ChatSessionKey(*sessionID)

	var session model.ChatSession
	if hit, _ := a.cache.Get(ctx, sessionKey, &session); !hit {
		session = model.ChatSession{
			SessionID:  *sessionID,
			HospitalID: hospitalID,
			UserID:     userID,
		}
	}

	session.History = append(session.History, llm.Message{
		Role:    llm.RoleUser,
		Content: message,
	})

	hospital, err := a.hospitalService.GetHospitalById(ctx, hospitalID)
	if err != nil {
		return nil, err
	}

	departments, err := a.departmentService.GetDepartments(ctx)
	if err != nil {
		return nil, err
	}

	var knownPatient *KnownPatientInfo
	if session.PatientName != "" {
		knownPatient = &KnownPatientInfo{
			Fullname: session.PatientName,
			Phone:    session.PatientPhone,
			Email:    session.PatientEmail,
		}
	}
	systemMsg := buildSystemPrompt(hospital.Name, departments, knownPatient)
	toolSpecs := a.toolSpecs()

	var replyText string
	for i := 0; i < maxToolIterations; i++ {
		messages := append([]llm.Message{systemMsg}, session.History...)

		resp, err := a.llm.Chat(ctx, llm.ChatRequest{
			Message: messages,
			Tools:   toolSpecs,
		})

		slog.Info("chat request", "messages", messages, "tools", toolSpecs)
		if err != nil {
			return nil, err
		}

		slog.Info("chat response", "message", resp)
		slog.Info("------Tool Calls", "tool_calls", resp.Message.ToolCalls)

		if len(resp.Message.ToolCalls) > 0 {
			session.History = append(session.History, llm.Message{
				Role:      llm.RoleAssistant,
				ToolCalls: resp.Message.ToolCalls,
			})

			for _, tc := range resp.Message.ToolCalls {
				result, toolErr := a.dispatchTool(ctx, &session, tc)
				if toolErr != nil {
					result = map[string]string{"error": toolErr.Error()}
				}
				resultJSON, _ := json.Marshal(result)
				slog.Info("tool call result", "result", resultJSON)

				session.History = append(session.History, llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tc.ID,
					Content:    string(resultJSON),
				})
			}
			continue
		}

		replyText = resp.Message.Content
		session.History = append(session.History, llm.Message{
			Role:    llm.RoleAssistant,
			Content: replyText,
		})
		break
	}

	if replyText == "" {
		replyText = "Xin lỗi, tôi chưa xử lý được yêu cầu này, bạn vui lòng thử lại."
	}

	if err := a.cache.Set(ctx, sessionKey, session, sessionTTL); err != nil {
		return nil, err
	}

	return &ChatReply{SessionID: *sessionID, Reply: replyText}, nil
}

type availableSlot struct {
	ShiftID   uint   `json:"shift_id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
}

func (a *AgentService) buildAvailableSlots(ctx context.Context, shifts []model.ScheduleShift) ([]availableSlot, error) {
	slots := make([]availableSlot, 0)
	for _, shift := range shifts {
		if shift.Status != model.ShiftScheduled || len(shift.Doctors) == 0 {
			continue
		}

		booked, err := a.appointmentService.ListBookedTimes(ctx, shift.ID)
		if err != nil {
			return nil, err
		}
		bookedSet := make(map[string]bool, len(booked))
		for _, t := range booked {
			if parsed, err := parseClockTime(t); err == nil {
				bookedSet[parsed.Format("15:04")] = true
			}
		}

		start, err := parseClockTime(shift.TimeStart)
		if err != nil {
			continue
		}
		end, err := parseClockTime(shift.TimeEnd)
		if err != nil {
			continue
		}

		for t := start; t.Before(end); t = t.Add(appointmentDuration) {
			hhmm := t.Format("15:04")
			if !bookedSet[hhmm] {
				slots = append(slots, availableSlot{
					ShiftID:   shift.ID,
					Date:      shift.Date.Format("2006-01-02"),
					StartTime: hhmm,
				})
			}
		}
	}
	return slots, nil
}

func (a *AgentService) dispatchTool(ctx context.Context, session *model.ChatSession, tc llm.ToolCall) (any, error) {
	switch tc.Name {
	case "search_available_slots":
		var args struct {
			DepartmentID *uint  `json:"department_id"`
			Date         string `json:"date"`
		}
		if err := json.Unmarshal(tc.Args, &args); err != nil {
			return nil, err
		}

		filter := dto.FilterScheduleShift{DepartmentID: args.DepartmentID}
		if args.Date != "" {
			parsed, err := time.Parse("2006-01-02", args.Date)
			if err != nil {
				return nil, fmt.Errorf("invalid date %q, expected YYYY-MM-DD", args.Date)
			}
			filter.Date = &parsed
		}

		session.DepartmentID = args.DepartmentID
		session.Date = args.Date

		shifts, _, err := a.scheduleService.FindScheduleShifts(ctx, session.HospitalID, filter)
		if err != nil {
			return nil, err
		}
		result, err := a.buildAvailableSlots(ctx, shifts)
		if err != nil {
			return nil, err
		}

		slog.Info("available slots", "slots", result)
		return result, nil

	case "propose_slot":
		var args struct {
			ShiftID   uint   `json:"shift_id"`
			StartTime string `json:"start_time"`
		}
		if err := json.Unmarshal(tc.Args, &args); err != nil {
			return nil, err
		}
		session.ShiftID = &args.ShiftID
		session.StartTime = args.StartTime
		session.PendingConfirmation = true
		return map[string]string{"status": "proposed"}, nil

	case "create_appointment":
		if !session.PendingConfirmation || session.ShiftID == nil || session.StartTime == "" {
			return nil, fmt.Errorf("chưa có slot nào được xác nhận trước đó")
		}

		slog.Info("create_appointment", "shift_id", *session.ShiftID, "start_time", session.StartTime)

		reason := model.ReasonNewPatientVisit
		if session.Reason != nil {
			reason = model.AppointmentReason(*session.Reason)
		}

		err := a.appointmentService.CreateAppointment(ctx, dto.CreateAppointmentInput{
			ShiftID:         *session.ShiftID,
			PatientName:     session.PatientName,
			PatientPhone:    session.PatientPhone,
			PatientEmail:    session.PatientEmail,
			AppointmentTime: session.StartTime,
			Reason:          reason,
		}, session.HospitalID)
		if err != nil {
			if err == apperror.ErrShiftAlreadyBooked {
				session.PendingConfirmation = false
				session.ShiftID = nil
			}
			slog.Error("create_appointment", "error", err)
			return nil, err
		}

		session.PendingConfirmation = false
		return map[string]string{"status": "booked"}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", tc.Name)
	}
}
