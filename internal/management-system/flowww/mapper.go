package flowww

import (
	"appointment-system/internal/domain"
	"fmt"
	"strings"
	"time"
)

type FlowwwMapper struct{}

func NewFlowwwMapper() *FlowwwMapper {
	return &FlowwwMapper{}
}

func (m *FlowwwMapper) MapFlowwwAppointmentDTOToAppointment(
	responseBody []FlowwwAppointmentDTO,
	centerId string,
) ([]domain.Appointment, error) {
	appmnts := make([]domain.Appointment, 0, len(responseBody))
	for _, responseItem := range responseBody {
		expirationDate, err := time.Parse("02/01/2006", responseItem.Date)
		if err != nil {
			return nil, err
		}

		expirationTimestamp := expirationDate.Unix() + 86400 // Add a day to the timestamp of startTime
		appmnts = append(appmnts, domain.Appointment{
			ID: m.BuildIdFromAppointmentIdAndCenterId(
				centerId, responseItem.ID,
			),
			CenterID:                        centerId,
			CustomerId:                      responseItem.CustomerId,
			CustomerName:                    responseItem.CustomerName,
			CustomerPhoneNumber:             responseItem.CustomerPhoneNumber,
			Date:                            responseItem.Date,
			StartTime:                       responseItem.StartTime,
			EndTime:                         responseItem.EndTime,
			Status:                          domain.AppointmentStatusPending,
			ConfirmationExpirationTimestamp: expirationTimestamp,
		})
	}

	return appmnts, nil
}

func (m *FlowwwMapper) BuildIdFromAppointmentIdAndCenterId(
	appointmentId, centerId string,
) string {
	return fmt.Sprintf("%s#%s", centerId, appointmentId)
}

func (m *FlowwwMapper) ExtractFlowwwAppointmentId(compositeId string) string {
	parts := strings.SplitN(compositeId, "#", 2)
	return parts[0]
}
