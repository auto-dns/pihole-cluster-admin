import { ReactNode } from 'react';
import classNames from 'classnames';
import styles from './AppCenteredPage.module.scss';

type Props = {
	children: ReactNode;
	className?: string;
};

export function AppCenteredPage({ className, children }: Props) {
	return <div className={classNames(className, styles.centeredPage)}>{children}</div>;
}
