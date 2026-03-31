export interface TaxReturn {
  id: string
  tax_year: number
  state_code: string
  filing_status: string | null
  created_at: string
  updated_at: string
}

export interface TaxReturnDetail extends TaxReturn {
  fields: FieldValue[]
}

export interface FieldValue {
  field_key: string
  value_num: number | null
  value_str: string | null
  source: "user_input" | "computed" | "prior_year" | "pdf_import"
}

export interface FormMeta {
  id: string
  name: string
  jurisdiction: string
  question_group: string
  question_order: number
  field_count: number
}

export interface FormDetail {
  id: string
  name: string
  jurisdiction: string
  question_group: string
  question_order: number
  fields: FieldMeta[]
}

export interface FieldMeta {
  line: string
  field_key: string
  field_type: "user_input" | "computed" | "lookup" | "prior_year" | "federal_ref"
  value_type: "numeric" | "string" | "integer"
  label: string
  prompt: string | null
  depends_on: string[]
  options: string[]
}

export interface RollforwardResponse {
  return_id: string
  tax_year: number
  fields_carried: number
  prior_year_values: number
}

export interface PriorYearValue {
  field_key: string
  source_year: number
  value_num: number | null
  value_str: string | null
}

export interface ChangedField {
  key: string
  value_num: number | null
  value_str: string | null
}

export interface UpdateFieldResponse {
  changed_fields: ChangedField[]
}

export interface ValidationResult {
  code: string
  severity: "error" | "warning" | "info"
  message: string
  field_key: string | null
}

export interface ValidateResponse {
  results: ValidationResult[]
  is_valid: boolean
}

export interface ExplainResponse {
  explanation: string
  model: string
  configured: boolean
}
