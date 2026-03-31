import { create } from "zustand"
import type { FormMeta, FormDetail } from "@/api/types"
import { fetchForms, fetchForm } from "@/api/hooks"

interface FormStore {
  forms: FormMeta[]
  formDetails: Map<string, FormDetail>
  loading: boolean

  loadForms: () => Promise<void>
  loadFormDetail: (formId: string) => Promise<FormDetail>
  getFormDetail: (formId: string) => FormDetail | undefined
}

export const useFormStore = create<FormStore>((set, get) => ({
  forms: [],
  formDetails: new Map(),
  loading: false,

  loadForms: async () => {
    if (get().forms.length > 0) return // already loaded
    set({ loading: true })
    try {
      const forms = await fetchForms()
      set({ forms, loading: false })
    } catch {
      set({ loading: false })
    }
  },

  loadFormDetail: async (formId) => {
    const cached = get().formDetails.get(formId)
    if (cached) return cached

    const detail = await fetchForm(formId)
    const newDetails = new Map(get().formDetails)
    newDetails.set(formId, detail)
    set({ formDetails: newDetails })
    return detail
  },

  getFormDetail: (formId) => get().formDetails.get(formId),
}))
