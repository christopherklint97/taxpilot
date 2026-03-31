mod f1040;
mod form_1116;
mod form_2555;
mod form_8833;
mod form_8889;
mod form_8938;
mod form_8949;
mod form_8995;
mod schedule_1;
mod schedule_2;
mod schedule_3;
mod schedule_a;
mod schedule_b;
mod schedule_c;
mod schedule_d;
mod schedule_se;

pub use f1040::form_1040;
pub use form_1116::form_1116;
pub use form_2555::form_2555;
pub use form_8833::form_8833;
pub use form_8889::form_8889;
pub use form_8938::form_8938;
pub use form_8949::form_8949;
pub use form_8995::form_8995;
pub use schedule_1::schedule_1;
pub use schedule_2::schedule_2;
pub use schedule_3::schedule_3;
pub use schedule_a::schedule_a;
pub use schedule_b::schedule_b;
pub use schedule_c::schedule_c;
pub use schedule_d::schedule_d;
pub use schedule_se::schedule_se;

use crate::domain::field::FormDef;

/// Returns all 16 federal form definitions.
pub fn all_federal_forms() -> Vec<FormDef> {
    vec![
        form_1040(),
        schedule_a(),
        schedule_b(),
        schedule_c(),
        schedule_d(),
        schedule_1(),
        schedule_2(),
        schedule_3(),
        schedule_se(),
        form_8949(),
        form_8889(),
        form_8995(),
        form_2555(),
        form_1116(),
        form_8938(),
        form_8833(),
    ]
}
