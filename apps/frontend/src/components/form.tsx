import type { InputHTMLAttributes } from "react";

type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
  name: string;
};

/**
 * @params label: the visible label text; name: the input name read on submit; rest: any other input attributes
 * @return a labelled input element
 * Renders one form field, wrapping the input in its label so no id wiring is needed.
 */
export function Field({ label, ...input }: FieldProps) {
  return (
    <label>
      {label}
      <input {...input} />
    </label>
  );
}

/**
 * @params error: the mutation error, or null when there is none
 * @return the error paragraph, or null when there is no error
 * Shows a mutation's error message under a form.
 */
export function FormError({ error }: { error: Error | null }) {
  if (!error) {
    return null;
  }
  return <p className="error">{error.message}</p>;
}

/**
 * @params form: the submitted form element; names: the field names to read
 * @return an object mapping each name to its string value ("" when the field is missing)
 * Reads the named fields of a form as strings, typed by the names asked for.
 */
export function readFields<K extends string>(form: HTMLFormElement, names: readonly K[]): Record<K, string> {
  const data = new FormData(form);
  const values = {} as Record<K, string>;
  for (const name of names) {
    values[name] = String(data.get(name) ?? "");
  }
  return values;
}
