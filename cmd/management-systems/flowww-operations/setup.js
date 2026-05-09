const { XMLParser } = require('fast-xml-parser');

const { REQUESTS_HEADERS, LOGIN_REQUEST } = require('./requests.js');

// Define the parser
const parser = new XMLParser({
    ignoreAttributes: false, // Necesario para leer d1="..." como "@_d1"
    attributeNamePrefix: "@_", // Así se distinguen atributos de valores
    parseTagValue: false, // Mantiene los valores como texto
    parseAttributeValue: false, // Evita convertir números a number
    htmlEntities: true, // Decodifica cosas como &#186;
    processEntities: {
        enabled: true,
        maxTotalExpansions: 10000,
    },
});

const apiConfig = {
    headers: null,
    baseRequestUrl: null,
    browser: null,
    page: null
};

async function Login(page) {
    // 1. Abrir login y completar credenciales
    await page.goto(process.env.FLOWWW_LOGIN_URL, { waitUntil: 'networkidle2' });
    await page.type(LOGIN_REQUEST.username_input_field, process.env.FLOWWW_USERNAME);
    await page.type(LOGIN_REQUEST.password_input_field, process.env.FLOWWW_PASSWORD);
    // 2. Enviar formulario y esperar navegación
    await Promise.all([
        page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 100000 }),
        page.click(LOGIN_REQUEST.submit_button),
    ]);
    console.log("✅ Login realizado");
}

const setupBrowser = async () => {
    const isLambda = !!process.env.AWS_LAMBDA_FUNCTION_NAME;
    const chromium = require('@sparticuz/chromium');
    const puppeteer = require('puppeteer-core');
    let browser;

    let executablePath;
    if (isLambda) {
        executablePath = await chromium.executablePath();
    } else {
        require('dotenv').config();
        executablePath = require('puppeteer').executablePath();
    }

    browser = await puppeteer.launch({
        args: chromium.args,
        defaultViewport: chromium.defaultViewport,
        executablePath: executablePath,
        headless: chromium.headless,
    });

    console.log(`✅ Navegador configurado. Entorno Lambda: ${isLambda}`);
    return browser;
}

const setupFlowwwConnection = async () => {
    const browser = await setupBrowser();
    const page = await browser.newPage();
    page.on('dialog', async dialog => {
        console.log('🪟 Dialog:', dialog.message());
        await new Promise(r => setTimeout(r, 500));
        await dialog.accept();
    });
    try {
        await Login(page);
    } catch (err) {
        throw new Error(`Error during FLOWww login: ${err.message}`);
    }

    const headers = await getRequestHeaders(page);
    const currentUrl = new URL(page.url());
    const basePath = currentUrl.pathname.substring(0, currentUrl.pathname.lastIndexOf('/'));
    const baseRequestUrl = `${currentUrl.origin}${basePath}/cgi-bin/dllrequest.asp`;
    apiConfig.browser = browser;
    apiConfig.page = page;
    apiConfig.headers = headers;
    apiConfig.baseRequestUrl = baseRequestUrl;
}

async function getRequestHeaders(page) {
    const cookies = await page.cookies();
    return {
        ...REQUESTS_HEADERS,
        'Cookie': cookies.map(c => `${c.name}=${c.value}`).join('; ')
    };
}

module.exports = {
    setupFlowwwConnection,
    apiConfig,
    parser,
};