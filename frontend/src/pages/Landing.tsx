import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Camera,
  PieChart,
  Sparkles,
  ArrowRight,
  Zap,
  ScanLine,
  BrainCircuit,
  BarChart3,
  Timer,
  Sun,
  Moon,
} from 'lucide-react';
import {
  ReceiptIllustration,
  BrainIllustration,
  ChartIllustration,
  RocketIllustration,
} from '@/components/Illustrations';
import { useTheme } from '@/context/ThemeContext';
import styles from './Landing.module.css';

const HERO_IMG = 'https://illustrations.popsy.co/violet/digital-nomad.svg';

const features = [
  {
    icon: <ScanLine size={26} />,
    title: 'Snap & Track',
    desc: 'Just take a photo of your receipt. Our AI reads it instantly and extracts every item.',
    color: 'var(--pink-dark)',
    bg: 'var(--pink-bg)',
    illustration: <ReceiptIllustration />,
  },
  {
    icon: <BrainCircuit size={26} />,
    title: 'AI Insights',
    desc: 'Get smart spending tips and warnings powered by multiple AI models.',
    color: 'var(--accent-dark)',
    bg: 'var(--accent-bg)',
    illustration: <BrainIllustration />,
  },
  {
    icon: <BarChart3 size={26} />,
    title: 'Visual Analytics',
    desc: 'Beautiful charts that make your spending patterns crystal clear.',
    color: 'var(--green-dark)',
    bg: 'var(--green-bg)',
    illustration: <ChartIllustration />,
  },
  {
    icon: <Timer size={26} />,
    title: 'Lightning Fast',
    desc: 'Upload to insights in seconds. No manual data entry needed.',
    color: 'var(--yellow-dark)',
    bg: 'var(--yellow-bg)',
    illustration: <RocketIllustration />,
  },
];

const steps = [
  {
    num: '01',
    icon: <Camera size={28} />,
    title: 'Snap a Receipt',
    desc: 'Take a photo or upload from gallery',
    color: 'var(--pink-dark)',
    bg: 'var(--pink-bg)',
  },
  {
    num: '02',
    icon: <Sparkles size={28} />,
    title: 'AI Does the Work',
    desc: 'Items extracted & categorized automatically',
    color: 'var(--accent-dark)',
    bg: 'var(--accent-bg)',
  },
  {
    num: '03',
    icon: <PieChart size={28} />,
    title: 'See Your Insights',
    desc: 'Charts, trends & smart spending tips',
    color: 'var(--green-dark)',
    bg: 'var(--green-bg)',
  },
];

export default function Landing() {
  const navigate = useNavigate();
  const { theme, toggle } = useTheme();

  return (
    <div className={styles.page}>
      {/* ── Decorative blobs ── */}
      <div className={styles.blobField} aria-hidden="true">
        <div className={`${styles.blob} ${styles.blob1}`} />
        <div className={`${styles.blob} ${styles.blob2}`} />
        <div className={`${styles.blob} ${styles.blob3}`} />
      </div>

      {/* ── Nav ── */}
      <nav className={styles.nav}>
        <motion.div
          className={styles.logo}
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <div className={styles.logoIcon}>
            <Sparkles size={18} />
          </div>
          <span className={styles.logoText}>Visora</span>
        </motion.div>
        <motion.div
          className={styles.navActions}
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
        >
          <button className={styles.themeBtn} onClick={toggle} aria-label="Toggle theme">
            {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
          </button>
          <button className={styles.navLink} onClick={() => navigate('/login')}>
            Log in
          </button>
          <button className={styles.navCta} onClick={() => navigate('/login')}>
            Get Started <ArrowRight size={16} />
          </button>
        </motion.div>
      </nav>

      {/* ── Hero ── */}
      <section className={styles.hero}>
        <div className={styles.heroContent}>
          <motion.div
            className={styles.heroBadge}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 }}
          >
            <Sparkles size={14} /> AI-Powered Expense Tracking
          </motion.div>

          <motion.h1
            className={styles.heroTitle}
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            Track Expenses
            <br />
            <span className="gradient-text">The Smart Way</span>
          </motion.h1>

          <motion.p
            className={styles.heroSub}
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.35 }}
          >
            Snap receipts, get AI-powered insights, and actually enjoy
            managing your money. No spreadsheets. No boring stuff.
          </motion.p>

          <motion.div
            className={styles.heroActions}
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.5 }}
          >
            <button className={styles.btnPrimary} onClick={() => navigate('/login')}>
              Start for Free <ArrowRight size={16} />
            </button>
            <button
              className={styles.btnSecondary}
              onClick={() =>
                document.getElementById('features')?.scrollIntoView({ behavior: 'smooth' })
              }
            >
              See How It Works
            </button>
          </motion.div>
        </div>

        <motion.div
          className={styles.heroVisual}
          initial={{ opacity: 0, scale: 0.9, x: 40 }}
          animate={{ opacity: 1, scale: 1, x: 0 }}
          transition={{ delay: 0.4, duration: 0.6 }}
        >
          <img
            src={HERO_IMG}
            alt="Person managing finances digitally"
            className={styles.heroIllustration}
          />

          <motion.div
            className={`${styles.floatCard} ${styles.floatCard1}`}
            animate={{ y: [0, -10, 0] }}
            transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
          >
            <ScanLine size={16} color="var(--accent-dark)" />
            <span>Receipt scanned</span>
          </motion.div>

          <motion.div
            className={`${styles.floatCard} ${styles.floatCard2}`}
            animate={{ y: [0, 8, 0] }}
            transition={{ duration: 3.5, repeat: Infinity, ease: 'easeInOut', delay: 0.5 }}
          >
            <Zap size={16} color="var(--green-dark)" />
            <span>24 items categorized</span>
          </motion.div>
        </motion.div>
      </section>

      {/* ── Features ── */}
      <section id="features" className={styles.features}>
        <motion.div
          className={styles.sectionHeader}
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <div className={styles.sectionIcon}>
            <Sparkles size={20} />
          </div>
          <h2>Why You'll Love It</h2>
          <p>Everything you need to master your spending, and actually enjoy it</p>
        </motion.div>

        <div className={styles.featureGrid}>
          {features.map((f, i) => (
            <motion.div
              key={f.title}
              className={styles.featureCard}
              initial={{ opacity: 0, y: 40 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              whileHover={{ y: -6, scale: 1.02 }}
            >
              <div className={styles.featureTop}>
                <div
                  className={styles.featureIcon}
                  style={{ background: f.bg, color: f.color }}
                >
                  {f.icon}
                </div>
                <div className={styles.featureIllustration}>
                  {f.illustration}
                </div>
              </div>
              <h3>{f.title}</h3>
              <p>{f.desc}</p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* ── How it works ── */}
      <section className={styles.howItWorks}>
        <motion.div
          className={styles.sectionHeader}
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <div className={styles.sectionIcon}>
            <Zap size={20} />
          </div>
          <h2>Three Steps. That's It.</h2>
          <p>From receipt to insights in under 10 seconds</p>
        </motion.div>

        <div className={styles.stepsRow}>
          {steps.map((s, i) => (
            <motion.div
              key={s.num}
              className={styles.stepCard}
              initial={{ opacity: 0, y: 40 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.15 }}
            >
              <div
                className={styles.stepIcon}
                style={{ background: s.bg, color: s.color }}
              >
                {s.icon}
              </div>
              <span className={styles.stepNum}>{s.num}</span>
              <h3>{s.title}</h3>
              <p>{s.desc}</p>
              {i < steps.length - 1 && (
                <div className={styles.stepConnector} aria-hidden="true">
                  <ArrowRight size={20} />
                </div>
              )}
            </motion.div>
          ))}
        </div>
      </section>

      {/* ── CTA ── */}
      <section className={styles.cta}>
        <motion.div
          className={styles.ctaInner}
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
        >
          <div className={styles.ctaIcon}>
            <Sparkles size={28} />
          </div>
          <h2>Ready to Take Control?</h2>
          <p>Join and start tracking your expenses the smart way</p>
          <button className={styles.btnPrimary} onClick={() => navigate('/login')}>
            Get Started (It's Free) <ArrowRight size={16} />
          </button>
        </motion.div>
      </section>

      {/* ── Footer ── */}
      <footer className={styles.footer}>
        <div className={styles.footerInner}>
          <div className={styles.footerBrand}>
            <div className={styles.logoIcon}><Sparkles size={14} /></div>
            <span className={styles.logoText}>Visora</span>
          </div>
          <p>© {new Date().getFullYear()} Visora. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
