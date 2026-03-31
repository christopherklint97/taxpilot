import { create } from "zustand"

interface UIStore {
  sidebarOpen: boolean
  selectedFormId: string | null
  toggleSidebar: () => void
  setSelectedForm: (formId: string | null) => void
}

export const useUIStore = create<UIStore>((set) => ({
  sidebarOpen: true,
  selectedFormId: null,
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
  setSelectedForm: (formId) => set({ selectedFormId: formId }),
}))
