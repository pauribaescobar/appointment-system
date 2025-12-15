const REQUESTS_HEADERS = {
      'Accept': '*/*',
      'Accept-Language': 'es-ES,es;q=0.6',
      'Connection': 'keep-alive',
      'Content-Type': 'application/x-www-form-urlencoded',
      'FLOWww-SessionID': '',
      'Origin': 'https://eu062.flowww.net',
      'Referer': 'https://eu062.flowww.net/sinvello/flowww.asp',
      'Sec-Fetch-Dest': 'empty',
      'Sec-Fetch-Mode': 'cors',
      'Sec-Fetch-Site': 'same-origin',
      'Sec-GPC': '1',
      'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36',
      'sec-ch-ua': '"Brave";v="137", "Chromium";v="137", "Not/A)Brand";v="24"',
      'sec-ch-ua-mobile': '?0',
      'sec-ch-ua-platform': '"macOS"'
};

const AGENDA_REQUEST_BODY ={
    'frm': 'SEC_DIARY_DAY_LOAD_FORM.V2',
    'SEC_DIARY_DAY_LOAD_FORM.V2_Fields': '2',
};

const LOGIN_REQUEST = {
    username_input_field: 'input[name="SEC_LOGIN_FORM_1"]',
    password_input_field:  'input[name="SEC_LOGIN_FORM_2"]',
    submit_button: 'flw-button'
}

const APPOINTMENT_CONFIRMATION_REQUEST_BODY = {
    'frm': 'SEC_APP_TAG_SAVE_FORM',
    'SEC_APP_TAG_SAVE_FORM_1':{
        "TagId": "1",
        "TagChecked":"-1"
    }
}

module.exports = {
    REQUESTS_HEADERS,
    AGENDA_REQUEST_BODY,
    LOGIN_REQUEST,
    APPOINTMENT_CONFIRMATION_REQUEST_BODY
};