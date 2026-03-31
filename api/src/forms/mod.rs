pub mod federal;
pub mod inputs;
pub mod state;

use crate::domain::field::FormDef;
use crate::domain::registry::Registry;

/// Registers all computed forms (federal + CA state) into the registry.
/// Input forms (W-2, 1099s) are NOT registered — their fields are provided
/// as runtime inputs via instance keys (e.g., "w2:1:wages").
pub fn register_all_forms() -> Registry {
    let mut registry = Registry::new();

    // Federal forms
    for form in federal::all_federal_forms() {
        registry.register(form);
    }

    // CA state forms
    registry.register(state::ca::f540::form_ca_540());
    registry.register(state::ca::schedule_ca::form_schedule_ca());
    registry.register(state::ca::form_3514::form_3514());
    registry.register(state::ca::form_3853::form_3853());

    registry
}

/// Returns all input form definitions (not registered in the solver).
pub fn all_input_forms() -> Vec<FormDef> {
    vec![
        inputs::w2::form_w2(),
        inputs::f1099_int::form_1099_int(),
        inputs::f1099_div::form_1099_div(),
        inputs::f1099_nec::form_1099_nec(),
        inputs::f1099_b::form_1099_b(),
    ]
}
