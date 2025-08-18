import { ReactNode } from 'react';
import styles from './AppCenteredPage.module.scss';
import classNames from 'classnames';

type Props = {
	children: ReactNode;
	className?: string;
};

export default function AppCenteredPage({ className, children }: Props) {
	return <div className={classNames(className, styles.centeredPage)}>{children}</div>;
}
