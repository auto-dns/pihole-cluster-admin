import { ChangeEvent } from 'react';
import { useState } from 'react';

export interface UseInputResult {
  value: string;
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
}

export default function useInput(initialValue: string): UseInputResult {
  const [value, setValue] = useState<string>(initialValue);

  function onChange(event: ChangeEvent<HTMLInputElement>) {
    setValue(event.target.value);
  }

  return { value, onChange };
}
