
const getAppointments = async (centerId) => {
    const { setupFlowwwConnection, apiConfig } = require('./setup.js');
    const { changeCenter, getAppointmentsFromFlowww } = require('./operations/get_appointments.js');

    await setupFlowwwConnection();
    await changeCenter(apiConfig.page, centerId);
    console.log(`✅ Cambiado al centro: ${centerId}`);
    const appointments = await getAppointmentsFromFlowww();
    console.log(`✅ Citas obtenidas: ${appointments.length}:`);
    await apiConfig.browser.close();

    return { statusCode: 200, body: JSON.stringify(appointments) };
}

const eventMapper = {
    "get-appointments":getAppointments
}

exports.handler = async (event, context) => {
    const operation = event.operation;
    const operationFunction = eventMapper[operation];
    if (operationFunction) {
        return await operationFunction(event.centerId);
    }

    return { statusCode: 200, body: JSON.stringify(false) };
};