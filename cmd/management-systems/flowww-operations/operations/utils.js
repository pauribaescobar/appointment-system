const transformDatefromMMDDYYYToDDMMYYYY = (date) => {
    const [month, day, year] = date.split('/');
    return `${day}/${month}/${year}`;
}

module.exports = {
    transformDatefromMMDDYYYToDDMMYYYY,
};