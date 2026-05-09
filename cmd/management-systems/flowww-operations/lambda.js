const { setupFlowwwConnection, apiConfig } = require('./setup.js');
const { changeCenter, getAppointmentsFromFlowww } = require('./operations/get_appointments.js');
const { confirmAppointment, removeAppointment } = require('./operations/update_confirmation_status.js');
const { toResponse, mapError, AppError } = require('./errors.js');

const getAppointments = async (event) => {
    try {
        if (!event?.centerId) {
            throw new AppError('INVALID_INPUT', 'MISSING_CENTER_ID', 'centerId is required');
        }

        await setupFlowwwConnection();

        await changeCenter(event.centerId);

        console.log(`✅ Cambiado al centro: ${event.centerId}`);

        const appointments = await getAppointmentsFromFlowww();

        console.log(`✅ Citas obtenidas: ${appointments.length}:`);

        return {
            statusCode: 200, body: JSON.stringify({
                ok: true,
                data: appointments
            })
        };
    } catch (err) {
        const appErr = err instanceof AppError ? err : mapError(err);
        console.error('❌ getAppointments error:', appErr);
        return toResponse(appErr);
    } finally {
        try {
            await apiConfig.browser.close();
        } catch (closeErr) {
            console.warn(`⚠️ Error al cerrar el navegador: ${closeErr.message}`);
        }
    }
}

const updateAppointmentStatusInputOk = (event) => {
    return event.centerId && (event.confirmationStatus === true || event.confirmationStatus === false) && event.appointmentDate && event.appointmentId;
}

const updateAppointmentStatus = async (event) => {
    const {
        centerId,
        confirmationStatus,
        appointmentDate,
        appointmentId,
    } = event;

    try {
        if (!updateAppointmentStatusInputOk(event)) {
            throw new AppError('INVALID_INPUT', 'MISSING_OR_INVALID_INPUT', 'centerId, confirmationStatus, appointmentDate and appointmentId are required');
        }

        await setupFlowwwConnection();
        await changeCenter(centerId);
        console.log(`✅ Cambiado al centro: ${centerId}`);
        const skipped = await (confirmationStatus ?
            confirmAppointment(appointmentId)
            : removeAppointment(appointmentId));
        return {
            statusCode: 200,
            body: JSON.stringify({ ok: true, data: { skipped: skipped } })
        };
    } catch (err) {
        const appErr = err instanceof AppError ? err : mapError(err);
        console.error('❌ updateAppointmentStatus error:', appErr);
        return toResponse(appErr);
    } finally {
        try {
            await apiConfig.browser.close();
        } catch (closeErr) {
            console.warn(`⚠️ Error al cerrar el navegador: ${closeErr.message}`);
        }
    }
}

const eventMapper = {
    "get-appointments": getAppointments,
    "confirmation-response": updateAppointmentStatus
}

exports.handler = async (event) => {
    const operation = event.operation;
    const operationFunction = eventMapper[operation];
    if (operationFunction) {
        return await operationFunction(event);
    }

    return { statusCode: 200, body: JSON.stringify(false) };
};


//For local testing purposes only

/*
(async () => {
    res = await getAppointments({
        operation: "get-appointments",
        centerId: "183"
    })

    //console.log(res.body);
    await updateAppointmentStatus({
        operation: "confirmation-response",
        centerId: "183",
        confirmationStatus: false,
        appointmentDate: "14/04/2026",
        appointmentId: "3995789"
    });
})();*/
