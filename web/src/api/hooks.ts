import { apiFetch } from "./client"
import type {
  TaxReturn,
  TaxReturnDetail,
  FormMeta,
  FormDetail,
  UpdateFieldResponse,
  RollforwardResponse,
  PriorYearValue,
  ValidateResponse,
  ExplainResponse,
} from "./types"

export async function fetchReturns(): Promise<TaxReturn[]> {
  return apiFetch("/returns")
}

export async function fetchReturn(id: string): Promise<TaxReturnDetail> {
  return apiFetch(`/returns/${id}`)
}

export async function createReturn(taxYear: number, stateCode?: string): Promise<TaxReturn> {
  return apiFetch("/returns", {
    method: "POST",
    body: JSON.stringify({ tax_year: taxYear, state_code: stateCode }),
  })
}

export async function deleteReturn(id: string): Promise<void> {
  return apiFetch(`/returns/${id}`, { method: "DELETE" })
}

export async function updateField(
  returnId: string,
  fieldKey: string,
  valueNum: number | null,
  valueStr: string | null,
): Promise<UpdateFieldResponse> {
  return apiFetch(`/returns/${returnId}/fields/${encodeURIComponent(fieldKey)}`, {
    method: "PUT",
    body: JSON.stringify({ value_num: valueNum, value_str: valueStr }),
  })
}

export async function rollforward(
  sourceReturnId: string,
  targetTaxYear?: number,
): Promise<RollforwardResponse> {
  return apiFetch(`/returns/${sourceReturnId}/rollforward`, {
    method: "POST",
    body: JSON.stringify({
      source_return_id: sourceReturnId,
      target_tax_year: targetTaxYear,
    }),
  })
}

export async function fetchPriorYearValues(
  returnId: string,
): Promise<PriorYearValue[]> {
  return apiFetch(`/returns/${returnId}/prior-year`)
}

export async function fetchForms(): Promise<FormMeta[]> {
  return apiFetch("/forms")
}

export async function fetchForm(formId: string): Promise<FormDetail> {
  return apiFetch(`/forms/${formId}`)
}

export async function validateReturn(returnId: string): Promise<ValidateResponse> {
  return apiFetch(`/returns/${returnId}/validate`)
}

export async function generateMefXml(returnId: string): Promise<Blob> {
  const res = await fetch(`/api/returns/${returnId}/efile/mef`, { method: "POST" })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.blob()
}

export async function generateCaXml(returnId: string): Promise<Blob> {
  const res = await fetch(`/api/returns/${returnId}/efile/ca`, { method: "POST" })
  if (!res.ok) throw new Error(`API error: ${res.status}`)
  return res.blob()
}

export function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement("a")
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export async function explainField(
  fieldKey: string,
  context?: string,
): Promise<ExplainResponse> {
  return apiFetch("/explain", {
    method: "POST",
    body: JSON.stringify({ field_key: fieldKey, context }),
  })
}
