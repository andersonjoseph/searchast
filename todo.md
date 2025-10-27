Here is a detailed unit testing plan for the specified methods in `context_builder.go`. 
The plan focuses on a "table-driven" testing methodology, which is idiomatic in Go for its clarity and maintainability.

### General Test Setup

Before diving into specific methods, you'll need a way to construct a `sourceTree` for your test cases. This will likely involve a test helper function. The `sourceTree` needs to model code with nested scopes.

**Example `sourceTree` Structure for Tests:**

Let's imagine a file structure like this to base our tests on:

```
1: func main() {
2:     // comment
3:     if true {
4:         for i := 0; i < 5; i++ {
5:             fmt.Println("Hello")
6:         }
7:     }
8:
9:     anotherFunc()
10: }
11:
12: func anotherFunc() {
13:     // another comment
14: }
```

This structure gives us:
*   A root scope (the whole file).
*   A parent scope `main` (lines 1-10).
*   A nested scope `if` (lines 3-7).
*   A deeply nested scope `for` (lines 4-6).
*   Another top-level scope `anotherFunc` (lines 12-14).

### 1. `addSurroundingLines` Method

**Objective:** To verify that the method correctly adds a specified number of lines before and after each line of interest, without going out of bounds.

**Test Cases (Table-Driven):**

| Test Name                 | `SurroundingLines` | Initial `linesOfInterest` | Expected `linesToShow`           | Rationale                                           |
| ------------------------- | ------------------ | ------------------------- | -------------------------------- | --------------------------------------------------- |
| **BasicCase**             | 2                  | `{5}`                     | `{3, 4, 5, 6, 7}`                | Tests the core functionality in a safe range.       |
| **AtStartOfFile**         | 3                  | `{1}`                     | `{1, 2, 3, 4}`                   | Ensures it doesn't try to add negative line numbers. |
| **AtEndOfFile**           | 3                  | `{13}`                    | `{10, 11, 12, 13, 14}`            | Ensures it doesn't exceed the file's line count.    |
| **OverlappingLines**      | 2                  | `{5, 8}`                  | `{3, 4, 5, 6, 7, 8, 9, 10}`      | Confirms that overlapping ranges are merged correctly. |
| **ZeroSurroundingLines**  | 0                  | `{5}`                     | `{5}`                            | Verifies that a value of 0 adds no context.         |
| **MultipleNonOverlapping**| 1                  | `{2, 13}`                 | `{1, 2, 3, 12, 13, 14}`          | Checks handling of multiple distinct lines.         |

