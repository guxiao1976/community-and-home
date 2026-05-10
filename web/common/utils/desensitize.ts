// Data desensitization utilities

/**
 * Desensitize phone number (138****0000)
 */
export function desensitizePhone(phone: string): string {
  if (!phone || phone.length !== 11) return phone;
  return phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2');
}

/**
 * Desensitize ID card number (110***********1234)
 */
export function desensitizeIdCard(idCard: string): string {
  if (!idCard || idCard.length !== 18) return idCard;
  return idCard.replace(/(\d{3})\d{11}(\d{4})/, '$1***********$2');
}

/**
 * Desensitize real name (张* or 欧阳**)
 */
export function desensitizeName(name: string): string {
  if (!name || name.length === 0) return name;
  if (name.length === 1) return name;
  return name.charAt(0) + '*'.repeat(name.length - 1);
}

/**
 * Desensitize email (abc***@example.com)
 */
export function desensitizeEmail(email: string): string {
  if (!email || !email.includes('@')) return email;
  const [username, domain] = email.split('@');
  if (username.length <= 3) {
    return username.charAt(0) + '***@' + domain;
  }
  return username.substring(0, 3) + '***@' + domain;
}
