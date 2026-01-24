# Bubbletea Key Reference

Complete reference for `KeyMsg.String()` values in Bubbletea's `tea.KeyMsg`.

[← Back to index](index.md)

## Basic Keys

| Key String    | Description              |
|---------------|--------------------------|
| `enter`       | Enter/Return key         |
| `tab`         | Tab key                  |
| `esc`         | Escape key               |
| `backspace`   | Backspace key            |
| `space`       | Space bar                |
| `delete`      | Delete key               |
| `insert`      | Insert key               |

## Navigation Keys

| Key String | Description        |
|------------|--------------------|
| `up`       | Up arrow           |
| `down`     | Down arrow         |
| `left`     | Left arrow         |
| `right`    | Right arrow        |
| `home`     | Home key           |
| `end`      | End key            |
| `pgup`     | Page Up            |
| `pgdown`   | Page Down          |

## Ctrl Combinations

### Ctrl + Letters

| Key String  | Common Binding    |
|-------------|-------------------|
| `ctrl+a`    | Select all        |
| `ctrl+b`    | Back character    |
| `ctrl+c`    | Interrupt/Copy    |
| `ctrl+d`    | Delete forward    |
| `ctrl+e`    | End of line       |
| `ctrl+f`    | Forward character |
| `ctrl+g`    | Cancel            |
| `ctrl+h`    | Backspace         |
| `ctrl+i`    | Tab (same as tab) |
| `ctrl+j`    | Newline           |
| `ctrl+k`    | Kill to end       |
| `ctrl+l`    | Clear screen      |
| `ctrl+m`    | Enter (same)      |
| `ctrl+n`    | Next line         |
| `ctrl+o`    | Open              |
| `ctrl+p`    | Previous line     |
| `ctrl+q`    | Quit              |
| `ctrl+r`    | Reverse search    |
| `ctrl+s`    | Save              |
| `ctrl+t`    | Transpose         |
| `ctrl+u`    | Kill to start     |
| `ctrl+v`    | Paste             |
| `ctrl+w`    | Kill word back    |
| `ctrl+x`    | Cut               |
| `ctrl+y`    | Yank              |
| `ctrl+z`    | Suspend           |

### Ctrl + Navigation

| Key String        | Description            |
|-------------------|------------------------|
| `ctrl+up`         | Ctrl + Up arrow        |
| `ctrl+down`       | Ctrl + Down arrow      |
| `ctrl+left`       | Ctrl + Left (word)     |
| `ctrl+right`      | Ctrl + Right (word)    |
| `ctrl+home`       | Ctrl + Home            |
| `ctrl+end`        | Ctrl + End             |
| `ctrl+pgup`       | Ctrl + Page Up         |
| `ctrl+pgdown`     | Ctrl + Page Down       |

### Ctrl + Special

| Key String        | Description            |
|-------------------|------------------------|
| `ctrl+@`          | Ctrl + @ (null)        |
| `ctrl+[`          | Escape (same as esc)   |
| `ctrl+]`          | Ctrl + ]               |
| `ctrl+^`          | Ctrl + ^               |
| `ctrl+_`          | Ctrl + _               |
| `ctrl+space`      | Ctrl + Space           |
| `ctrl+backspace`  | Ctrl + Backspace       |

## Shift Combinations

| Key String       | Description           |
|------------------|-----------------------|
| `shift+tab`      | Shift + Tab (backtab) |
| `shift+up`       | Shift + Up            |
| `shift+down`     | Shift + Down          |
| `shift+left`     | Shift + Left          |
| `shift+right`    | Shift + Right         |
| `shift+home`     | Shift + Home          |
| `shift+end`      | Shift + End           |
| `shift+pgup`     | Shift + Page Up       |
| `shift+pgdown`   | Shift + Page Down     |

## Ctrl+Shift Combinations

| Key String            | Description              |
|-----------------------|--------------------------|
| `ctrl+shift+up`       | Ctrl + Shift + Up        |
| `ctrl+shift+down`     | Ctrl + Shift + Down      |
| `ctrl+shift+left`     | Ctrl + Shift + Left      |
| `ctrl+shift+right`    | Ctrl + Shift + Right     |
| `ctrl+shift+home`     | Ctrl + Shift + Home      |
| `ctrl+shift+end`      | Ctrl + Shift + End       |

## Function Keys

| Key String | Key String  |
|------------|-------------|
| `f1`       | `f11`       |
| `f2`       | `f12`       |
| `f3`       | `f13`       |
| `f4`       | `f14`       |
| `f5`       | `f15`       |
| `f6`       | `f16`       |
| `f7`       | `f17`       |
| `f8`       | `f18`       |
| `f9`       | `f19`       |
| `f10`      | `f20`       |

## Alt Modifier

Alt modifies the key string by prepending `alt+`:

| Pattern              | Example        |
|----------------------|----------------|
| `alt+<letter>`       | `alt+x`        |
| `alt+<number>`       | `alt+1`        |
| `alt+<nav>`          | `alt+up`       |
| `alt+<function>`     | `alt+f1`       |
| `alt+enter`          | Alt + Enter    |
| `alt+backspace`      | Alt + Backspace|

## Usage Example

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            m.cursor--
        case "down", "j":
            m.cursor++
        case "enter":
            m.selected = m.cursor
        }
    }
    return m, nil
}
```

## Notes

- Key strings are **lowercase** (use `ctrl+c` not `Ctrl+C`)
- Multiple keys can match the same action: `ctrl+m` = `enter`, `ctrl+i` = `tab`
- Printable characters return themselves: `a`, `1`, `@`, etc.
- Unknown keys return the raw rune value
