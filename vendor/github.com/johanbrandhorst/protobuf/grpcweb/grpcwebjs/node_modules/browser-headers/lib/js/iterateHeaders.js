// This function is written in JS (ES5) to avoid an issue with TypeScript targeting ES5, but requiring Symbol.iterator
function iterateHeaders(headers, callback) {
  var iterator = headers[Symbol.iterator]();
  var entry = iterator.next();
  while(!entry.done) {
    callback(entry.value[0]);
    entry = iterator.next();
  }
}

function iterateHeadersKeys(headers, callback) {
  var iterator = headers.keys();
  var entry = iterator.next();
  while(!entry.done) {
    callback(entry.value);
    entry = iterator.next();
  }
}

module.exports = {
  iterateHeaders: iterateHeaders,
  iterateHeadersKeys: iterateHeadersKeys
};
