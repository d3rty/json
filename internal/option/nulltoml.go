package option

// TomlNone stands for a magic string "None" that allows use to specify option.None() in toml files
// Beucase of TOML has no concept of null and both BurntSushi / pelletier parsers can't handle empty tables well
// we'll workout with such a hack
const TomlNone = "None"
