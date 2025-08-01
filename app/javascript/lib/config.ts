export interface SiteConfig {
  branding: {
    name: string;
    logo: string;
    link: string;
  };
  theme: {
    primary: string;
    secondary: string;
  };
  hero: {
    title: string;
    subtitle: string;
    description: string;
    learnMoreText: string;
    learnMoreUrl: string;
  };
  footer: {
    brandName: string;
    link: string;
  };
}

export const siteConfig: SiteConfig = {
  branding: {
    name: "Takiacademy",
    logo: "https://play-lh.googleusercontent.com/QwR54xVSABVpntrh-PkcP37uzjUE2wv4o1YoGDCD4ISrVhHTQegWohH-GDppGjQVWr7h",
    link: "https://takiacademy.com",
  },
  theme: {
    primary: "#35bbe3",
    secondary: "100 116 139",
  },
  hero: {
    title: "Welcome to BigBlueButton, hello hazem",
    subtitle: "Transform Your Online Learning Experience",
    description:
      "BigBlueButton is an open source web conferencing system for online classes. The platform maximizes time for applied learning by enabling students to collaborate and receive feedback in real-time.",
    learnMoreText: "Learn more about BigBlueButton",
    learnMoreUrl: "#",
  },
  footer: {
    brandName: "Greenlight",
    link: "https://docs.bigbluebutton.org/greenlight/v3/install",
  },
};
