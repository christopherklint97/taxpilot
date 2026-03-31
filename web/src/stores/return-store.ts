import { create } from "zustand"
import type { TaxReturn } from "@/api/types"
import { apiFetch } from "@/api/client"

interface ReturnStore {
  returns: TaxReturn[]
  loading: boolean
  error: string | null

  fetchReturns: () => Promise<void>
  createReturn: (taxYear: number, stateCode?: string) => Promise<TaxReturn>
  deleteReturn: (id: string) => Promise<void>
}

export const useReturnStore = create<ReturnStore>((set, get) => ({
  returns: [],
  loading: false,
  error: null,

  fetchReturns: async () => {
    set({ loading: true, error: null })
    try {
      const returns = await apiFetch<TaxReturn[]>("/returns")
      set({ returns, loading: false })
    } catch (e) {
      set({ error: (e as Error).message, loading: false })
    }
  },

  createReturn: async (taxYear, stateCode) => {
    const ret = await apiFetch<TaxReturn>("/returns", {
      method: "POST",
      body: JSON.stringify({ tax_year: taxYear, state_code: stateCode }),
    })
    set({ returns: [ret, ...get().returns] })
    return ret
  },

  deleteReturn: async (id) => {
    await apiFetch(`/returns/${id}`, { method: "DELETE" })
    set({ returns: get().returns.filter((r) => r.id !== id) })
  },
}))
