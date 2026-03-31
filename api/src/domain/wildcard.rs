/// Matches a key against a pattern that may contain `*` wildcards.
/// Each `*` matches zero or more characters.
///
/// This matches the Go implementation exactly: the pattern is split on `*`,
/// and segments must appear in order. The first segment must match at the start,
/// the last segment must match at the end.
///
/// Examples:
///   match_wildcard("w2:*:wages", "w2:1:wages") => true
///   match_wildcard("1040:line1", "1040:line1") => true
///   match_wildcard("w2:*:wages", "1099int:1:interest") => false
pub fn match_wildcard(pattern: &str, s: &str) -> bool {
    let parts: Vec<&str> = pattern.split('*').collect();
    if parts.len() == 1 {
        return pattern == s;
    }

    let mut idx = 0usize;
    for (i, part) in parts.iter().enumerate() {
        if part.is_empty() {
            continue;
        }
        if let Some(pos) = s[idx..].find(part) {
            if i == 0 && pos != 0 {
                // First part must match at start
                return false;
            }
            idx += pos + part.len();
        } else {
            return false;
        }
    }
    // If the last part is non-empty, the string must end with it
    let last = parts[parts.len() - 1];
    if !last.is_empty() {
        return s.ends_with(last);
    }
    true
}

/// Given a value_pattern, a filter_pattern, and a concrete key that matched
/// value_pattern, returns the concrete key that corresponds to filter_pattern
/// with the same wildcard segment.
///
/// Example:
///   build_corresponding_key("1099b:*:proceeds", "1099b:*:term", "1099b:1:proceeds")
///   => Some("1099b:1:term")
pub fn build_corresponding_key(
    value_pattern: &str,
    filter_pattern: &str,
    concrete_key: &str,
) -> Option<String> {
    let v_parts: Vec<&str> = value_pattern.splitn(2, '*').collect();
    if v_parts.len() != 2 {
        return None; // only single-wildcard patterns supported
    }
    let prefix = v_parts[0];
    let suffix = v_parts[1];

    if !concrete_key.starts_with(prefix) || !concrete_key.ends_with(suffix) {
        return None;
    }

    let wildcard_seg = &concrete_key[prefix.len()..concrete_key.len() - suffix.len()];
    Some(filter_pattern.replace('*', wildcard_seg))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_match_wildcard_exact() {
        assert!(match_wildcard("1040:line1", "1040:line1"));
        assert!(!match_wildcard("1040:line1", "1040:line2"));
    }

    #[test]
    fn test_match_wildcard_star() {
        assert!(match_wildcard("w2:*:wages", "w2:1:wages"));
        assert!(match_wildcard("w2:*:wages", "w2:2:wages"));
        assert!(match_wildcard("w2:*:wages", "w2::wages")); // * matches zero chars in Go
        assert!(!match_wildcard("w2:*:wages", "1099int:1:interest"));
    }

    #[test]
    fn test_build_corresponding_key() {
        assert_eq!(
            build_corresponding_key("1099b:*:proceeds", "1099b:*:term", "1099b:1:proceeds"),
            Some("1099b:1:term".to_string())
        );
        assert_eq!(
            build_corresponding_key("w2:*:wages", "w2:*:state_wages", "w2:abc:wages"),
            Some("w2:abc:state_wages".to_string())
        );
    }
}
