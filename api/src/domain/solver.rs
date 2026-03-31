use std::collections::{BTreeSet, HashMap, HashSet};

use crate::domain::field::{match_wildcard, DepValues, FieldType};
use crate::domain::registry::Registry;

/// DependencyGraph builds and resolves the DAG of all form fields.
pub struct DependencyGraph<'a> {
    registry: &'a Registry,
    /// Adjacency list: key -> list of keys it depends on.
    edges: HashMap<String, Vec<String>>,
    /// All known field keys.
    nodes: HashSet<String>,
}

impl<'a> DependencyGraph<'a> {
    /// Creates a new DependencyGraph backed by the given registry.
    pub fn new(registry: &'a Registry) -> Self {
        Self {
            registry,
            edges: HashMap::new(),
            nodes: HashSet::new(),
        }
    }

    /// Constructs the dependency graph from all registered forms.
    /// Returns an error if a cycle is detected.
    pub fn build(&mut self) -> Result<(), String> {
        self.edges.clear();
        self.nodes.clear();

        // First pass: register all nodes and expand wildcards against known nodes so far
        for form in self.registry.all_forms() {
            for field in &form.fields {
                let key = format!("{}:{}", form.id, field.line);
                self.nodes.insert(key.clone());

                let mut resolved = Vec::new();
                for dep in &field.depends_on {
                    if dep.contains('*') {
                        // Expand wildcard to all matching registered field keys
                        for node in &self.nodes {
                            if match_wildcard(dep, node) {
                                resolved.push(node.clone());
                            }
                        }
                    } else {
                        resolved.push(dep.clone());
                    }
                }
                self.edges.insert(key, resolved);
            }
        }

        // Second pass: expand wildcards again now that all nodes are known
        for form in self.registry.all_forms() {
            for field in &form.fields {
                let key = format!("{}:{}", form.id, field.line);
                for dep in &field.depends_on {
                    if dep.contains('*') {
                        let existing: HashSet<String> = self
                            .edges
                            .get(&key)
                            .map(|v| v.iter().cloned().collect())
                            .unwrap_or_default();

                        let mut new_deps = Vec::new();
                        for node in &self.nodes {
                            if match_wildcard(dep, node) && !existing.contains(node) {
                                new_deps.push(node.clone());
                            }
                        }
                        if let Some(edge_list) = self.edges.get_mut(&key) {
                            edge_list.extend(new_deps);
                        }
                    }
                }
            }
        }

        // Check for cycles via topological sort
        self.topological_sort()?;
        Ok(())
    }

    /// Returns all field keys in dependency order using Kahn's algorithm.
    /// Fields with no dependencies come first.
    pub fn topological_sort(&self) -> Result<Vec<String>, String> {
        // Compute in-degree for each node
        let mut in_degree: HashMap<&str, usize> = HashMap::new();
        for node in &self.nodes {
            in_degree.insert(node.as_str(), 0);
        }

        // Build reverse adjacency: dep -> list of keys that depend on it
        let mut reverse_adj: HashMap<&str, Vec<&str>> = HashMap::new();
        for (key, deps) in &self.edges {
            for dep in deps {
                // Only count edges to known nodes
                if self.nodes.contains(dep) {
                    *in_degree.entry(key.as_str()).or_insert(0) += 1;
                    reverse_adj
                        .entry(dep.as_str())
                        .or_default()
                        .push(key.as_str());
                }
            }
        }

        // Start with nodes that have zero in-degree (sorted for determinism)
        let mut queue: BTreeSet<&str> = BTreeSet::new();
        for (node, &deg) in &in_degree {
            if deg == 0 {
                queue.insert(node);
            }
        }

        let mut sorted = Vec::new();
        while let Some(node) = queue.iter().next().copied() {
            queue.remove(node);
            sorted.push(node.to_string());

            if let Some(dependents) = reverse_adj.get(node) {
                for &dependent in dependents {
                    if let Some(deg) = in_degree.get_mut(dependent) {
                        *deg -= 1;
                        if *deg == 0 {
                            queue.insert(dependent);
                        }
                    }
                }
            }
        }

        if sorted.len() != self.nodes.len() {
            // Find nodes involved in cycle for error message
            let mut cycle_nodes: Vec<&str> = in_degree
                .iter()
                .filter(|&(_, deg)| *deg > 0)
                .map(|(&node, _)| node)
                .collect();
            cycle_nodes.sort();
            return Err(format!(
                "cycle detected involving fields: {}",
                cycle_nodes.join(", ")
            ));
        }

        Ok(sorted)
    }

    /// Returns all UserInput field keys that don't have values in the provided map.
    pub fn missing_inputs(&self, provided: &HashMap<String, f64>) -> Vec<String> {
        let mut missing = Vec::new();
        for form in self.registry.all_forms() {
            for field in &form.fields {
                if field.field_type == FieldType::UserInput {
                    let key = format!("{}:{}", form.id, field.line);
                    if !provided.contains_key(&key) {
                        missing.push(key);
                    }
                }
            }
        }
        missing.sort();
        missing
    }

    /// Resolves all Computed fields given the provided UserInput values.
    /// Returns all field values (inputs + computed).
    pub fn solve(
        &self,
        inputs: &HashMap<String, f64>,
        str_inputs: &HashMap<String, String>,
        tax_year: i32,
    ) -> Result<HashMap<String, f64>, String> {
        // Check for missing required inputs
        let missing = self.missing_inputs(inputs);
        if !missing.is_empty() {
            return Err(format!(
                "missing required inputs: {}",
                missing.join(", ")
            ));
        }

        let order = self.topological_sort()?;

        // Initialize result with provided inputs
        let mut result: HashMap<String, f64> = inputs.clone();
        let mut str_result: HashMap<String, String> = str_inputs.clone();

        // Process fields in topological order
        for key in &order {
            if result.contains_key(key) {
                continue; // already provided as input
            }

            let (_form, field) = match self.registry.get_field(key) {
                Ok(v) => v,
                Err(_) => continue,
            };

            match field.field_type {
                FieldType::UserInput => {
                    // Should already be in result
                    continue;
                }
                FieldType::Computed
                | FieldType::Lookup
                | FieldType::FederalRef
                | FieldType::PriorYear => {
                    let dv = DepValues::new(result.clone(), str_result.clone(), tax_year);
                    if let Some(ref compute) = field.compute {
                        result.insert(key.clone(), compute(&dv));
                    }
                    if let Some(ref compute_str) = field.compute_str {
                        str_result.insert(key.clone(), compute_str(&dv));
                    }
                }
            }
        }

        Ok(result)
    }
}
