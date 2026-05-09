const { apiConfig } = require('../setup.js');

const CANCELATION_REASON_CONFIRMATION_NOT_ACCEPTED = "CITA ELIMINADA. HA DICHO QUE NO ASISTIRÁ";

const clickConfirmAppointmentButton = async (appointmentId) => {
    const payload = JSON.stringify({ DiaryGID: String(appointmentId), TagID: "1", TagChecked: "-1" });
    const response = await fetch(apiConfig.baseRequestUrl, {
        method: 'POST',
        headers: apiConfig.headers,
        body: `frm=SEC_APP_TAG_SAVE_FORM&SEC_APP_TAG_SAVE_FORM_1=${encodeURIComponent(payload)}`
    });

    if (!response.ok) {
        throw new Error(`Failed to confirm appointment ${appointmentId}: HTTP ${response.status}`);
    }

    return true;
}

const clickNotAssistedAppointmentButton = async (appointmentId) => {
    const response = await fetch(apiConfig.baseRequestUrl, {
        method: 'POST',
        headers: apiConfig.headers,
        body: `frm=SEC_APP_NOT_ATTENDED_FORM&SEC_APP_NOT_ATTENDED_FORM_Fields=3&SEC_APP_NOT_ATTENDED_FORM_1=${appointmentId}&SEC_APP_NOT_ATTENDED_FORM_2=1&SEC_APP_NOT_ATTENDED_FORM_3=${encodeURIComponent(CANCELATION_REASON_CONFIRMATION_NOT_ACCEPTED)}`
    });

    if (!response.ok) {
        throw new Error(`Failed to mark appointment ${appointmentId} as not attended: HTTP ${response.status}`);
    }
}

const confirmAppointment = async (appointmentId) => {
    try {
        await clickConfirmAppointmentButton(appointmentId);
        console.log(`✅ Cita ${appointmentId} confirmada`);
        return true; // appointment confirmed
    } catch (error) {
        console.error(`❌ Error al confirmar cita: ${error.message}`);
        throw error;
    }
}

const removeAppointment = async (appointmentId) => {
    try {
        await clickNotAssistedAppointmentButton(appointmentId);
        console.log(`✅ Cita ${appointmentId} eliminada`);
        return true; // appointment removed
        //return false; // appointment not found, not removed, skipped
    } catch (error) {
        console.error(`❌ Error al eliminar cita: ${error.message}`);
        throw error;
    }
}

module.exports = {
    confirmAppointment,
    removeAppointment
}