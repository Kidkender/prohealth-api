package event

type Event interface {
	Name() string
}

type AppointmentCreated struct {
	AppointmentID uint
	ActorID       *uint
	PatientName   string
	PatientEmail  string
	Reason        string
	Date          string
	Time          string
}

func (AppointmentCreated) Name() string { return "appointment.created" }
