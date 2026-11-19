## Root cause

`LoadConfigStores` decodes each layer into its own value, runs `Normalize` on it, then folds the layers with `later.Merge(earlier)`. A field a layer's file omits is zero in that layer's value, so `Merge` cannot tell "this file did not set it" from "this file set it to zero", and anything `Normalize` fills into a later layer reads as a value that layer set. Defaults therefore hide earlier layers' real values, and no layer can set `false`, `0` or `""` over an earlier one.

## Fix: layered decode

`encoding/json/v2` leaves a struct field untouched when the JSON being decoded omits its key. Decoding every layer's file, in order, into one value therefore gives "later layer wins only for the keys it has" with no presence tracking. Verified in a scratch program on Go 1.27:

- a later `"verbose":false` replaced an earlier `true`; omitted keys kept earlier values;
- maps merged; slices were replaced;
- a later `"theme":null` reset the field to zero;
- a slice type with its own `UnmarshalJSONFrom` appended instead of replacing, which is how a field overrides the default.

Defaults applied to the fresh value before the first layer decodes survive every layer that omits the key, so the E-2661 bug cannot occur.

## Decisions (Mike, in the /needs-plan exchange)

- Q1 = B: `ApplyDefaults(DefaultsArgs) error` is a required `RootConfig` method, so a developer sees where defaults go. Making it optional later, if it proves burdensome, would not break implementers.
- Q2 = ii: defaults are written into a CLI config file go-cfgstore creates. E-2663 covers telling users when a written default goes stale.
- Q3 = a (layered decode) over `Opt[T]`: `Opt[T]` needs `.Value()` on every read, `OptOf` in every literal, and an untyped `cfgstore:"append"` struct tag for per-field merge rules; Mike objected to the untyped tag. Layered decode keeps plain fields, and per-field overrides are typed (`AppendSlice[T]`, `ReplaceMap[K,V]`).
- `Merge` stays a required method (Mike: a developer needs a method to implement as the place to start writing code). Its common-case body is one line, `return args.Decode(c)`, and it takes `MergeArgs[*RC]`, so there is no type assertion.
- Q4: the failed-load wrap names E-2658 (corrupt-config detection and repair); E-2658 and E-2659 get notes pointing here.
- Q6 = i: layered decode is E-2661's fix rather than a separate task; an after-Merge defaults hook would be rewritten as soon as layered decode landed.

## Normalize sees the accumulated value

With layered decode, the `Normalize` run after each layer sees every layer decoded so far. This is not a cost for Endless or Gomion:

- Endless's `Normalize` validates opt-in values and names `SourceFile`. Earlier layers' values were validated after their own layer, and validation moves out of `Normalize` into the check pass (see below), which sees each file alone.
- Gomion's `Normalize` stamps `Schema` and `Version`, which is idempotent.

It would only matter for a `Normalize` that is not idempotent (appending, or rewriting a value it already rewrote), so the doc comment requires idempotence.

## Project config init

`InitProjectConfig` does not write defaults into the new project file. A project file holding every default would override each user's CLI-layer settings, which is the E-2661 bug again in a different place. The config it returns has defaults applied in memory, after the save.

## Reporting every error at once (Q7 = D′)

`encoding/json/v2` stops decoding a file at its first invalid field (verified: a field type whose `UnmarshalJSONFrom` rejected a value hid a second invalid field and a type error in the same file). Validating during decode therefore reports one error per file. Validating in `Normalize` during the layered pass would recheck earlier layers' values and blame later files. So validation runs in a check pass over each file decoded alone: go-cfgstore walks the value and calls `Validate` on each field type implementing `FieldValidator`, then calls the app's `RootConfig.Validate` for cross-field checks, and joins every error from every file. Endless already does this by hand (`OptInText` holds the raw string, `walkFields` walks the struct, `validateOptIns` joins the errors); under this design `OptInText` gets a `Validate` method and the walker moves into go-cfgstore. Rejected: stopping at the first failing layer (whack-a-mole across files); running `Normalize` on each file alone as well as layered (`Normalize` would see two kinds of value); an app-written `Validate` only (each app writes its own walk).
