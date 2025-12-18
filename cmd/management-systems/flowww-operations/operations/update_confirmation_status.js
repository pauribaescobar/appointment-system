const { APPOINTMENT_CONFIRMATION_REQUEST_BODY, REQUESTS_HEADERS } = require('../requests');
const qs = require('qs');
const { apiConfig } = require('../setup.js');

const CANCELATION_REASON_CONFIRMATION_NOT_ACCEPTED = "CITA ELIMINADA. HA DICHO QUE NO ASISTIRÁ";

const transformDatefromDDMMYYYYToMMDDYYYY = (date) => {
    const [day, month, year] = date.split('/');
    return `${month}/${day}/${year}`;
}

const navigateToAgendaModule = async () => {
    const page = apiConfig.page;
    try {
         console.log("✅ Navegando a la agenda");
        await page.waitForSelector('flw-menu-item[id="MAIN.DIARY"]');
        await page.evaluate(() => {
            agendaConatiner = document.querySelector('flw-menu-item[id="MAIN.DIARY"]');
            agendaButton = agendaConatiner.querySelector('.content');
            agendaButton.click();
        });
        console.log("✅ Módulo de agenda abierto");
        await page.reload();
    }catch (error) {
        console.error(`❌ Error al navegar al módulo de agenda: ${error.message}`);
        throw error;
    }
}

const goToAppointmentDateAgendaPage = async (appointmentDate) => {
    const page = apiConfig.page;
    try {
        await page.waitForSelector('li[id=day]');
        const appointmentDateObj = new Date(transformDatefromDDMMYYYYToMMDDYYYY(appointmentDate)).getTime();
        let currentDate = new Date(transformDatefromDDMMYYYYToMMDDYYYY(
            await page.evaluate(() => {
                return document.querySelector('li#day').innerText;
            })
        )).getTime();
        while(currentDate != appointmentDateObj){
            const currentDateValue = await page.evaluate(async () => {
                const navBarDay = document.querySelector('div.navBarDay');
                const nextButton = (navBarDay.querySelector('li.diary__nav_spanleft')).querySelector('a');
                nextButton.click();
                await new Promise(r => setTimeout(r, 2000));
                return document.querySelector('li#day').innerText;
            });
            currentDate = new Date(transformDatefromDDMMYYYYToMMDDYYYY(currentDateValue)).getTime();
        }
        console.log(`✅ Día de la cita ${appointmentDate} cargado`);
    }catch (error) {
        console.error(`❌ Error al navegar a la fecha de la cita: ${error.message}`);
        throw error;
    }
}

const loadAppointment = async (appointmentId) => {
    const page = apiConfig.page;
    try {
        const selector = `div[diaryid="${appointmentId}"]`;
        await page.waitForSelector(selector, { timeout: 90000 });
        await page.click(selector, { clickCount: 2 });
        console.log(`✅ Cita ${appointmentId} abierta`);

        await new Promise(r => setTimeout(r, 5000));
    } catch (error) {
        console.error(`❌ Error al cargar la cita: ${error.message}`);
        throw error;
    }
}

const clickConfirmAppointmentButton = async () => {
    const page = apiConfig.page;
    try {
        await page.evaluate(async ()=>{
            const tagsButton = document.querySelector('button#btnTagsColors');
            tagsButton.click();
            await new Promise(r => setTimeout(r, 7500));
            const confirmAppointmentButton = document.querySelector('li[style="background-color: rgb(76, 175, 80);"] label input');
            if (confirmAppointmentButton.checked) return;
            console.log("✅ Confirmando cita...");
            confirmAppointmentButton.click();
        })
    } catch (error) {
        console.error(`❌ Error al hacer click en el botón de confirmar cita: ${error.message}`);
        throw error;
    }
}

const clickNotAssistedAppointmentButton = async () => {
    const page = apiConfig.page;
    try {
        page.removeAllListeners('dialog');
        page.on('dialog', async dialog => {
            await dialog.accept(CANCELATION_REASON_CONFIRMATION_NOT_ACCEPTED);
        });
        await page.evaluate(async ()=>{
            const deleteButton = document.querySelector('button#SEC_APP_SAVE_FORM_Bt7');
            deleteButton.click();
        })
    } catch (error) {
        console.error(`❌ Error al hacer click en el botón de eliminar cita: ${error.message}`);
        throw error;
    }
}

const navigatetoAppointmentsPage = async (appointmentId, appointmentDate) => {
    try {
        // Ir al módulo de agenda
        await navigateToAgendaModule();
        // Ir al día de la cita
        await goToAppointmentDateAgendaPage(appointmentDate);
        // Hacer doble click en la cita
        await loadAppointment(appointmentId); 
    } catch (error) {
        console.error(`❌ Error al navegar a la página de citas: ${error.message}`);
        throw error;
    }
}

const confirmAppointment = async (appointmentDate, appointmentId) => {
    try {
        // Navegar a la página de la cita
        await navigatetoAppointmentsPage(appointmentId, appointmentDate);
        // Hacer click en el botón de confirmar cita
        await clickConfirmAppointmentButton();
        console.log(`✅ Cita ${appointmentId} confirmada`);
    } catch (error) {
        console.error(`❌ Error al confirmar cita: ${error.message}`);
        throw error;
    }
}

const removeAppointment = async (appointmentDate, appointmentId) => {
    try {
        // Navegar a la página de la cita
        await navigatetoAppointmentsPage(appointmentId, appointmentDate);
        // Hacer click en el botón de cita no asistida
        await clickNotAssistedAppointmentButton();
        console.log(`✅ Cita ${appointmentId} eliminada`);
    } catch (error) {
        console.error(`❌ Error al eliminar cita: ${error.message}`);
        throw error;
    }
    return;
}

module.exports = {
    confirmAppointment,
    removeAppointment
}