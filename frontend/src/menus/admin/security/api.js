// Fungsi menu Security Center.
import { apiRequest as request } from '../../../api/client.js'

export async function securitySummary() {
  return request('/api/admin/security/summary', { auth: true })
}
