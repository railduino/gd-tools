//
// Local functions for MeetingPilot
//

function crc32(str) {
  const table = new Uint32Array(256).map((_, i) => {
    for (let j = 0; j < 8; j++) {
      i = (i & 1 ? 0xEDB88320 : 0) ^ (i >>> 1);
    }
    return i >>> 0;
  });
  let crc = 0xFFFFFFFF;
  for (let i = 0; i < str.length; i++) {
    crc = (crc >>> 8) ^ table[(crc ^ str.charCodeAt(i)) & 0xFF];
  }
  let result = (crc ^ 0xFFFFFFFF) >>> 0;
  return result.toString(16).padStart(8, "0").toUpperCase();
}

function checkValue(value, list) {
  const crcValue = crc32(value);
  return list.includes(crcValue);
}

