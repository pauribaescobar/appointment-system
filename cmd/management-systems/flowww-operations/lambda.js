const { setupFlowwwConnection, apiConfig } = require('./setup.js');
const { changeCenter, getAppointmentsFromFlowww } = require('./operations/get_appointments.js');
const { confirmAppointment, removeAppointment } = require('./operations/update_confirmation_status.js');

const getAppointments = async (event) => {
    await setupFlowwwConnection();
    await changeCenter(apiConfig.page, event.centerId);
    console.log(`✅ Cambiado al centro: ${event.centerId}`);
    const appointments = await getAppointmentsFromFlowww();
    console.log(`✅ Citas obtenidas: ${appointments.length}:`);
    await apiConfig.browser.close();

    return { statusCode: 200, body: JSON.stringify(appointments) };
}

const updateAppointmentStatus = async (event) => {
    const {
        centerId,
        confirmationStatus,
        appointmentDate,
        appointmentId,
    } = event;

    await setupFlowwwConnection();
    await changeCenter(apiConfig.page, centerId);
    console.log(`✅ Cambiado al centro: ${centerId}`);

    await new Promise(r => setTimeout(r, 5000));
    const result = await (confirmationStatus ? 
        confirmAppointment(appointmentDate, appointmentId)
        : removeAppointment(appointmentDate, appointmentId))
    ;
    await apiConfig.browser.close();

    return { statusCode: 200, body: JSON.stringify(result) };
}

const eventMapper = {
    "get-appointments":getAppointments,
    "confirmation-response":updateAppointmentStatus
}

exports.handler = async (event) => {
    const operation = event.operation;
    const operationFunction = eventMapper[operation];
    if (operationFunction) {
        return await operationFunction(event);
    }

    return { statusCode: 200, body: JSON.stringify(false) };
};

/**
 * For local testing purposes only
 */
/** 
(async ()=> {
    await getAppointments({
        operation:"get-appointments",
        centerId: "183"
    })
    await updateAppointmentStatus({
        operation:"confirmation-response",
        centerId: "183",
        confirmationStatus:false,
        appointmentDate:"23/12/2025",
        appointmentId: "3267653"
    });
})();
*/