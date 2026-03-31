use crate::domain::field::*;
use crate::domain::form::*;

pub fn form_8949() -> FormDef {
    FormDef {
        id: FORM_F8949.to_string(),
        name: "Form 8949 — Sales and Other Dispositions of Capital Assets".to_string(),
        jurisdiction: Jurisdiction::Federal,
        tax_years: vec![2024, 2025, 2026],
        question_group: "income_1099".to_string(),
        question_order: 3,
        fields: vec![
            // --- Part I: Short-Term ---
            // Total short-term proceeds
            {
                let deps = vec![
                    F1099_B_WILDCARD_PROCEEDS.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("st_proceeds", "Short-term total proceeds", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_PROCEEDS, F1099_B_WILDCARD_TERM, "short")
                }))
            },
            // Total short-term cost basis
            {
                let deps = vec![
                    F1099_B_WILDCARD_BASIS.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("st_basis", "Short-term total cost basis", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_BASIS, F1099_B_WILDCARD_TERM, "short")
                }))
            },
            // Total short-term wash sale adjustments
            {
                let deps = vec![
                    F1099_B_WILDCARD_WASH_SALE.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("st_wash", "Short-term wash sale adjustments", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_WASH_SALE, F1099_B_WILDCARD_TERM, "short")
                }))
            },
            // Short-term gain or loss
            {
                let deps = vec![
                    F8949_ST_PROCEEDS.to_string(),
                    F8949_ST_BASIS.to_string(),
                    F8949_ST_WASH.to_string(),
                ];
                FieldDef::new_computed("st_gain_loss", "Short-term gain or (loss)", deps, Box::new(|dv: &DepValues| {
                    dv.get(F8949_ST_PROCEEDS) - dv.get(F8949_ST_BASIS) + dv.get(F8949_ST_WASH)
                }))
            },
            // --- Part II: Long-Term ---
            // Total long-term proceeds
            {
                let deps = vec![
                    F1099_B_WILDCARD_PROCEEDS.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("lt_proceeds", "Long-term total proceeds", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_PROCEEDS, F1099_B_WILDCARD_TERM, "long")
                }))
            },
            // Total long-term cost basis
            {
                let deps = vec![
                    F1099_B_WILDCARD_BASIS.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("lt_basis", "Long-term total cost basis", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_BASIS, F1099_B_WILDCARD_TERM, "long")
                }))
            },
            // Total long-term wash sale adjustments
            {
                let deps = vec![
                    F1099_B_WILDCARD_WASH_SALE.to_string(),
                    F1099_B_WILDCARD_TERM.to_string(),
                ];
                FieldDef::new_computed("lt_wash", "Long-term wash sale adjustments", deps, Box::new(|dv: &DepValues| {
                    dv.sum_all_where(F1099_B_WILDCARD_WASH_SALE, F1099_B_WILDCARD_TERM, "long")
                }))
            },
            // Long-term gain or loss
            {
                let deps = vec![
                    F8949_LT_PROCEEDS.to_string(),
                    F8949_LT_BASIS.to_string(),
                    F8949_LT_WASH.to_string(),
                ];
                FieldDef::new_computed("lt_gain_loss", "Long-term gain or (loss)", deps, Box::new(|dv: &DepValues| {
                    dv.get(F8949_LT_PROCEEDS) - dv.get(F8949_LT_BASIS) + dv.get(F8949_LT_WASH)
                }))
            },
        ],
    }
}
