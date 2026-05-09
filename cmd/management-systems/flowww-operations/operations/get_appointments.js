const qs = require('qs');
const { apiConfig, parser } = require('../setup.js');
const { transformDatefromMMDDYYYToDDMMYYYY } = require('./utils.js');

const { AGENDA_REQUEST_BODY } = require('../requests.js');

const getChangeCenterRequestBody = (centerId) => {
  return `frm=SEC_CONFIG_CLISEL_FORM.V4&SEC_CONFIG_CLISEL_FORM.V4_Fields=1&SEC_CONFIG_CLISEL_FORM.V4_1=${centerId}`;
}

const refreshPage = async () => {
  await fetch(apiConfig.baseRequestUrl, {
    method: 'POST',
    headers: apiConfig.headers,
    body: `frm=SEC_REFRESH_FORM&SEC_REFRESH_FORM_Fields=0`
  });
}

// FASTEST WAY TO CHANGE CENTER
async function changeCenter(centerId) {
  await fetch(apiConfig.baseRequestUrl, {
    method: 'POST',
    headers: apiConfig.headers,
    body: getChangeCenterRequestBody(centerId)
  });
  await refreshPage();
}

const buildGetAppointmentsRequestBody = (date) => {
  return qs.stringify({
    ...AGENDA_REQUEST_BODY,
    'SEC_DIARY_DAY_LOAD_FORM.V2_1': date.toISOString().split('T')[0].split('-').reverse().join('/') // DD/MM/YYYY
  })
}

function getConfirmationDate() {
  const date = new Date();
  date.setDate(date.getDate() + 2);
  return date;
}

const parseAppointmentData = async (response) => {
  const textResponse = await response.text();
  const jsonResponse = (parser.parse(textResponse)).FLOWww_AES.FLOWww_Info.FLOWww_List[0].FLOWww_Item;
  return (jsonResponse ? jsonResponse.map(item => ({
    id: item['@_d1'], // Hora
    customerId: item['@_d20'], // Identificador del Cliente
    customerName: item['@_d9'].split(" (n")[0], // Nombre del Cliente
    customerPhoneNumber: item['@_d10'], // Teléfono
    date: item['@_d3'] ? transformDatefromMMDDYYYToDDMMYYYY(item['@_d3']) : '', // Fecha del tratamiento en formato DD/MM/YYYY
    startTime: item['@_d4'], // Hora en que empieza el tratamiento
    endTime: item['@_d5'], // Hora en que termina el tratamiento
    inAgenda: item['@_d18'] === '0' ? true : false, // Si la cita está en la agenda
  })) : []).filter(appointment =>
    appointment.customerId &&
    appointment.customerName &&
    appointment.customerPhoneNumber &&
    appointment.inAgenda
  ); // Filtrar citas no asignadas
}

const getAppointmentsFromFlowww = async () => {
  try {
    const confirmationDate = getConfirmationDate();
    const response = await fetch(apiConfig.baseRequestUrl, {
      method: 'POST',
      headers: apiConfig.headers,
      body: buildGetAppointmentsRequestBody(confirmationDate)
    });
    const appointments = await parseAppointmentData(response);
    return appointments;
  } catch (error) {
    console.error(`❌ Error al obtener citas: ${error.message}`);
    throw error;
  }
}

module.exports = {
  changeCenter,
  getAppointmentsFromFlowww
};