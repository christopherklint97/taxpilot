import { create } from "zustand"

interface UIStore {
  sidebarOpen: boolean
  selectedFormId: string | null
  /** Show only user-input fields (hide computed) */
  inputsOnly: boolean
  /** Forms the user has explicitly hidden */
  hiddenForms: Set<string>
  /** Show all forms including empty ones */
  showEmptyForms: boolean

  toggleSidebar: () => void
  setSelectedForm: (formId: string | null) => void
  toggleInputsOnly: () => void
  toggleFormHidden: (formId: string) => void
  toggleShowEmptyForms: () => void
}

export const useUIStore = create<UIStore>((set) => ({
  sidebarOpen: true,
  selectedFormId: null,
  inputsOnly: false,
  hiddenForms: new Set(),
  showEmptyForms: false,

  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSelectedForm: (formId) => set({ selectedFormId: formId }),
  toggleInputsOnly: () => set((s) => ({ inputsOnly: !s.inputsOnly })),
  toggleFormHidden: (formId) =>
    set((s) => {
      const next = new Set(s.hiddenForms)
      if (next.has(formId)) next.delete(formId)
      else next.add(formId)
      return { hiddenForms: next }
    }),
  toggleShowEmptyForms: () => set((s) => ({ showEmptyForms: !s.showEmptyForms })),
}))
