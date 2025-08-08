import { PiholeNode } from '../../types/pihole';
import useInput from '../../hooks/useInput';

export default function PiholeNodeForm({ node }: { node: PiholeNode }) {
	const [scheme, setScheme] = useInput();
	return <form></form>;
}
