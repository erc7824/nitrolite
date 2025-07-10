import React from 'react';
import Link from '@docusaurus/Link';
import clsx from 'clsx';
import styles from './styles.module.css';

export function Card({ title, description, to, icon }) {
  return (
    <Link 
      to={to} 
      className={styles.card}
    >
      <div className={styles.cardContent}>
        {icon && <div className={styles.cardIcon}>{icon}</div>}
        <div className={styles.cardHeader}>
          <h3 className={styles.cardTitle}>{title}</h3>
        </div>
        {description && (
          <p className={styles.cardDescription}>{description}</p>
        )}
      </div>
    </Link>
  );
}

export function CardGrid({ children, cols = 2 }) {
  return (
    <div className={clsx(styles.cardGrid, styles[`cols${cols}`])}>
      {children}
    </div>
  );
}