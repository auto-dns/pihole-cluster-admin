import { PiholeNode } from '@/types/pihole';
import { useClusterHealth } from '@/hooks/useClusterHealth';
import { PiholeStatusLight } from '@/components/StatusLight/PiholeStatusLight';
import styles from './PiholeCardList.module.scss';

type Props = {
	nodes: PiholeNode[];
	onCardClick: (node: PiholeNode) => void;
};

export function PiholeCardList({ nodes, onCardClick }: Props) {
	const { nodeHealthById, isFresh } = useClusterHealth();

	return (
		<ul className={styles.cardList} role='list' aria-label='Configured Pi-hole nodes'>
			{nodes.map((node) => {
				const health = nodeHealthById.get(node.id);
				const url = `${node.scheme}://${node.host}:${node.port}`;

				return (
					<li key={node.id} className={styles.item}>
						<button
							type='button'
							className={styles.cardButton}
							onClick={() => onCardClick(node)}
							aria-label={`Edit ${node.name}`}
							title='Tap to edit'
						>
							<div className={styles.header}>
								<PiholeStatusLight
									name={node.name}
									health={health}
									fresh={isFresh}
								/>
								<span className={styles.name}>{node.name}</span>
								<span className={styles.chevron} aria-hidden='true'>
									›
								</span>
							</div>

							<div className={styles.row}>
								<span className={styles.label}>URL</span>
								<span className={styles.url} title={url}>
									{url}
								</span>
							</div>

							<div className={styles.row}>
								<span className={styles.label}>Description</span>
								<span className={styles.desc} title={node.description || '-'}>
									{node.description || '-'}
								</span>
							</div>
						</button>
					</li>
				);
			})}
		</ul>
	);
}
