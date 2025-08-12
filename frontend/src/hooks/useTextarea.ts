import { ChangeEvent } from 'react';
import { useState } from 'react';

export interface UseInputResult {
	value: string;
	onChange: (event: ChangeEvent<HTMLTextAreaElement>) => void;
}

export default function useInput(initialValue: string): UseInputResult {
	const [value, setValue] = useState<string>(initialValue);

	function onChange(event: ChangeEvent<HTMLTextAreaElement>) {
		setValue(event.target.value);
	}

	return { value, onChange };
}
