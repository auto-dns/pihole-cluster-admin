import { ReactNode } from 'react';
import classNames from 'classnames';
import styles from './AppCard.module.scss';

type Props = {
	children: ReactNode;
	className?: string;
};

export function AppCard({ className, children }: Props) {
	return <div className={classNames(className, styles.card)}>{children}</div>;
}
