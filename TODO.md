Refactoring the `SourceTree` struct to decouple the context-building logic is an excellent way to improve the code's design. This will make the code more modular, easier to test, and more maintainable in the long run. Here is a comprehensive plan to achieve this refactoring:

### 1. Analysis of the Current `SourceTree` Struct

First, let's analyze the current responsibilities of the `SourceTree` struct. It is responsible for:

*   **Data Storage**: It holds the parsed source code (`lines`), lines of interest (`linesOfInterest`), and lines to be displayed (`linesToShow`).
*   **Parsing and Tree Building**: The `NewSourceTree` and `build` methods are responsible for parsing the source code and building the tree structure.
*   **Line Finding**: The `findLines` method searches for lines that match a given pattern.
*   **Context Building**: The `addContext`, `addParentContext`, `addChildContext`, and `closeGaps` methods are responsible for adding context to the lines of interest.
*   **Output Formatting**: The `formatOutput` method is responsible for formatting the final output.

The key issue is that the context-building logic is tightly coupled with the `SourceTree` struct. This makes it difficult to change the context-building strategy without modifying the `SourceTree` struct itself.

### 2. The Refactoring Plan: Introducing a `ContextBuilder`

To decouple the context-building logic, we will introduce a new `ContextBuilder` struct. This struct will be responsible for taking a `SourceTree` and a set of lines of interest and then adding the necessary context to the `linesToShow` set.

Here is the step-by-step refactoring plan:

**Step 1: Create the `ContextBuilder` Struct**

First, create a new `ContextBuilder` struct. This struct will hold the parameters that control the context-building process, such as the gap size for surrounding lines and the gap size for closing gaps.

```go
type ContextBuilder struct {
    GapToShow         lineNumber
    GapToClose        lineNumber
    ParentContext     bool
    ChildContext      bool
    SurroundingLines  bool
}
```

**Step 2: Create a `NewContextBuilder` Function**

Next, create a constructor function for the `ContextBuilder`. This function will initialize the `ContextBuilder` with default values for the context-building parameters.

```go
func NewContextBuilder() *ContextBuilder {
    return &ContextBuilder{
        GapToShow:         3,
        GapToClose:        3,
        ParentContext:     true,
        ChildContext:      true,
        SurroundingLines:  true,
    }
}
```

**Step 3: Move Context-Building Methods to the `ContextBuilder`**

Now, move the `addContext`, `addParentContext`, `addChildContext`, and `closeGaps` methods from the `SourceTree` struct to the `ContextBuilder`. These methods will now take a `*SourceTree` as a parameter.

```go
func (cb *ContextBuilder) AddContext(st *SourceTree) {
    // ...
}

func (cb *ContextBuilder) addParentContext(st *SourceTree, line lineNumber) {
    // ...
}

func (cb *ContextBuilder) addChildContext(st *SourceTree, line lineNumber) {
    // ...
}

func (cb *ContextBuilder) closeGaps(st *SourceTree) {
    // ...
}
```

**Step 4: Modify the `SourceTree` Struct**

With the context-building logic moved to the `ContextBuilder`, the `SourceTree` struct can be simplified. It will no longer need to hold the context-building parameters.

```go
type SourceTree struct {
    lines           []line
    linesOfInterest Set[lineNumber]
    linesToShow     Set[lineNumber]
    seenParents     Set[lineNumber]
}
```

**Step 5: Update the `main` Function**

Finally, update the `main` function to use the new `ContextBuilder`.

```go
func main() {
    sourceTree, err := NewSourceTree("./main.go")
    if err != nil {
        panic(err)
    }

    if _, err = sourceTree.findLines("AI\\?"); err != nil {
        panic(err)
    }

    contextBuilder := NewContextBuilder()
    contextBuilder.AddContext(&sourceTree)

    print(sourceTree.formatOutput())
}
```

### 3. Benefits of the Refactoring

This refactoring provides several benefits:

*   **Decoupling**: The `SourceTree` struct is now decoupled from the context-building logic. This means you can change the context-building strategy without modifying the `SourceTree` struct.
*   **Modularity**: The context-building logic is now encapsulated in the `ContextBuilder` struct, making the code more modular and easier to understand.
*   **Testability**: The `ContextBuilder` can be tested in isolation, making it easier to write unit tests for the context-building logic.
*   **Flexibility**: You can create different `ContextBuilder` configurations to achieve different context-building behaviors.

By following this refactoring plan, you can significantly improve the design of your code, making it more robust, maintainable, and easier to work with in the long run.
