const qs = require('qs');
const { apiConfig, parser } = require('../setup.js');

const { AGENDA_REQUEST_BODY } = require('../requests.js');

async function changeCenter(page, centerId) {
  try {
    // Esperar a que aparezcan los elementos
    await page.waitForSelector('flw-clinic-item');
    await new Promise(resolve => setTimeout(resolve, 7500));
    await page.evaluate(() => 
      {
        document.querySelector('flw-clinic-item').shadowRoot.querySelector('.clinic-item').click();
      });
    await new Promise(resolve => setTimeout(resolve, 7500));
    await page.evaluate((centerId) =>{
      const clinicItems = document.querySelectorAll('flw-clinic-item');
      let centerItem = null;
      clinicItems.forEach(clinicItem => {
        if(clinicItem.id === centerId){
          centerItem = clinicItem;
        }
      });
      centerItem.shadowRoot.querySelector('.clinic-item').click();
    }, centerId); 
  } catch (error) {
    console.error(`❌ Error al cambiar centro: ${error.message}`);
    throw error;
  }
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

const transformDatefromMMDDYYYToDDMMYYYY = (date) => {
    const [month, day, year] = date.split('/');
    return `${day}/${month}/${year}`;
}

const parseAppointmentData = async (response) => {
    const textResponse = await response.text();
    const jsonResponse = (parser.parse(textResponse)).FLOWww_AES.FLOWww_Info.FLOWww_List[0].FLOWww_Item;
    return jsonResponse.map(item =>({
        id: item['@_d1'], // Hora
        customerId: item['@_d20'], // Identificador del Cliente
        customerName: item['@_d9'].split(" (n")[0], // Nombre del Cliente
        customerPhoneNumber: item['@_d10'], // Teléfono
        date: item['@_d3'] ? transformDatefromMMDDYYYToDDMMYYYY(item['@_d3']) : '', // Fecha del tratamiento en formato DD/MM/YYYY
        startTime: item['@_d4'], // Hora en que empieza el tratamiento
        endTime: item['@_d5'] // Hora en que termina el tratamiento
    })).filter(appointment => 
        appointment.customerId &&
        appointment.customerName &&
        appointment.customerPhoneNumber
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