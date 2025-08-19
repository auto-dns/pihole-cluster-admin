import { ReactNode } from 'react';
import styles from './index.module.scss';
import classNames from 'classnames';

type Props = {
	children: ReactNode;
	className?: string;
};

export default function AppCard({ className, children }: Props) {
	return <div className={classNames(className, styles.card)}>{children}</div>;
}
